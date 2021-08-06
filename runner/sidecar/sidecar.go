package sidecar

import (
	"context"
	"crypto/tls"
	"fmt"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger              = sharedutil.NewLogger()
	clusterName         = os.Getenv(dfv1.EnvClusterName)
	namespace           = os.Getenv(dfv1.EnvNamespace)
	patchMu             = sync.Mutex{}
	pipelineName        = os.Getenv(dfv1.EnvPipelineName)
	ready               = false // we are ready to serve HTTP requests, also updates pod status condition
	dynamicInterface    dynamic.Interface
	lastStep            dfv1.Step
	kubernetesInterface kubernetes.Interface
	secretInterface     corev1.SecretInterface
	prePatchHooks       []func(ctx context.Context) error // hooks to run before patching
	replica             int
	step                dfv1.Step // this is updated on start, and then periodically as we update the status
	stepName            string
	updateInterval      time.Duration
)

func becomeUnreadyHook(context.Context) error {
	ready = false
	return nil
}

func Exec(ctx context.Context) error {
	restConfig := ctrl.GetConfigOrDie()
	dynamicInterface = dynamic.NewForConfigOrDie(restConfig)
	kubernetesInterface = kubernetes.NewForConfigOrDie(restConfig)
	secretInterface = kubernetesInterface.CoreV1().Secrets(namespace)

	sharedutil.MustUnJSON(os.Getenv(dfv1.EnvStep), &step)

	logger.Info("step", "clusterName", clusterName, "step", sharedutil.MustJSON(step))

	if clusterName == "" {
		// this must be configured in the controller
		return fmt.Errorf("cluster name (%q) was not specifcied", dfv1.EnvClusterName)
	}

	stepName = step.Spec.Name
	if step.Status.SourceStatuses == nil {
		step.Status.SourceStatuses = dfv1.SourceStatuses{}
	}
	if step.Status.SinkStatues == nil {
		step.Status.SinkStatues = dfv1.SourceStatuses{}
	}
	lastStep = *step.DeepCopy()

	if v, err := strconv.Atoi(os.Getenv(dfv1.EnvReplica)); err != nil {
		return err
	} else {
		replica = v
	}
	if v, err := time.ParseDuration(os.Getenv(dfv1.EnvUpdateInterval)); err != nil {
		return err
	} else {
		updateInterval = v
	}

	if err := enrichSpec(ctx); err != nil {
		return err
	}

	logger.Info("generating self-signed certificate")
	const certFile, keyFile = "/tmp/runner.crt", "/tmp/runner.key"
	if err := generateCert(certFile, keyFile); err != nil {
		return fmt.Errorf("failed to generate cert: %w", err)
	}

	logger.Info("sidecar config", "stepName", stepName, "pipelineName", pipelineName, "replica", replica, "updateInterval", updateInterval.String())

	defer logger.Info("done")
	defer stop()
	defer preStop("defer")

	addStopHook(patchStepStatusHook)

	toSinks, err := connectSinks(ctx)
	if err != nil {
		return err
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if ready {
			w.WriteHeader(204)
		} else {
			w.WriteHeader(503)
		}
	})
	addPreStopHook(becomeUnreadyHook)

	if leadReplica() {
		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "replicas",
			Help: "Number of replicas, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#replicas",
		}, func() float64 { return float64(step.Status.Replicas) })
		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "version_major",
			Help: "Major version number, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#version_major",
		}, func() float64 { return float64(sharedutil.Version.Major()) })
		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "version_minor",
			Help: "Minor version number, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#version_minor",
		}, func() float64 { return float64(sharedutil.Version.Minor()) })
		promauto.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "version_patch",
			Help: "Patch version number, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#version_patch",
		}, func() float64 { return float64(sharedutil.Version.Patch()) })
	}

	// we listen to this message, but it does not come from Kubernetes, it actually comes from the main container's
	// pre-stop hook
	http.HandleFunc("/pre-stop", func(w http.ResponseWriter, r *http.Request) {
		preStop(r.URL.Query().Get("source"))
		w.WriteHeader(204)
	})

	connectOut(toSinks)

	server := &http.Server{Addr: "localhost:3569"}
	addStopHook(func(ctx context.Context) error {
		logger.Info("closing HTTP server")
		return server.Shutdown(context.Background())
	})
	go func() {
		defer runtimeutil.HandleCrash()
		logger.Info("starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "failed to listen-and-server on HTTP")
		}
		logger.Info("HTTP server shutdown")
	}()
	httpServer := &http.Server{Addr: ":3570", TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	addStopHook(func(ctx context.Context) error {
		logger.Info("closing HTTPS server")
		return httpServer.Shutdown(context.Background())
	})
	go func() {
		defer runtimeutil.HandleCrash()
		logger.Info("starting HTTPS server")
		if err := httpServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "failed to listen-and-server on HTTPS")
		}
		logger.Info("HTTPS server shutdown")
	}()

	toMain, err := connectIn(ctx, toSinks)
	if err != nil {
		return err
	}

	if err := connectSources(ctx, toMain); err != nil {
		return err
	}

	go wait.JitterUntil(func() { defer runtimeutil.HandleCrash(); patchStepStatus() }, updateInterval, 1.2, true, ctx.Done())

	ready = true
	logger.Info("ready")
	<-ctx.Done()
	return nil
}

func patchStepStatusHook(context.Context) error {
	patchStepStatus()
	return nil
}

func patchStepStatus() {
	ctx := context.Background()
	patchMu.Lock()
	defer patchMu.Unlock()
	if ready { // don't run these if we are not ready, pollutes the logs with errors
		for _, f := range prePatchHooks {
			n := sharedutil.GetFuncName(f)
			logger.Info("executing pre-patch hook", "func", n)
			if err := f(ctx); err != nil {
				logger.Error(err, "failed to execute hook", "func", n)
			}
		}
	}
	if notEqual, patch := sharedutil.NotEqual(dfv1.Step{Status: lastStep.Status}, dfv1.Step{Status: step.Status}); notEqual {
		logger.Info("patching step status", "patch", patch)
		if un, err := dynamicInterface.
			Resource(dfv1.StepGroupVersionResource).
			Namespace(namespace).
			Patch(
				ctx,
				pipelineName+"-"+stepName,
				types.MergePatchType,
				[]byte(patch),
				metav1.PatchOptions{},
				"status",
			); err != nil {
			if !apierr.IsNotFound(err) { // the step can be deleted before the pod
				logger.Error(err, "failed to patch step status")
			}
		} else {
			v := dfv1.Step{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, &v); err != nil {
				logger.Error(err, "failed to from-unstructured")
			} else {
				lastStep = *v.DeepCopy()
				withLock(func() {
					if v.Status.SourceStatuses == nil {
						v.Status.SourceStatuses = dfv1.SourceStatuses{}
					}
					if v.Status.SinkStatues == nil {
						v.Status.SinkStatues = dfv1.SourceStatuses{}
					}
					// the step with change while this goroutine is running, so we must copy the data for this
					// replica back to the status
					r := strconv.Itoa(replica)
					for name, s := range step.Status.SourceStatuses {
						if v.Status.SourceStatuses[name].Metrics == nil {
							x := v.Status.SourceStatuses[name]
							x.Metrics = map[string]dfv1.Metrics{}
							v.Status.SourceStatuses[name] = x
						}
						v.Status.SourceStatuses[name].Metrics[r] = s.Metrics[r]
					}
					for name, s := range step.Status.SinkStatues {
						if v.Status.SinkStatues[name].Metrics == nil {
							x := v.Status.SinkStatues[name]
							x.Metrics = map[string]dfv1.Metrics{}
							v.Status.SinkStatues[name] = x
						}
						v.Status.SinkStatues[name].Metrics[r] = s.Metrics[r]
					}
					step = v
				})
			}
		}
	}
}

func enrichSpec(ctx context.Context) error {
	if err := enrichSources(ctx); err != nil {
		return err
	}
	return enrichSinks(ctx)
}

func enrichSources(ctx context.Context) error {
	for i, source := range step.Spec.Sources {
		if x := source.STAN; x != nil {
			if err := enrichSTAN(ctx, x); err != nil {
				return err
			}
			source.STAN = x
		} else if x := source.Kafka; x != nil {
			if err := enrichKafka(ctx, &x.Kafka); err != nil {
				return err
			}
			source.Kafka = x
		} else if x := source.S3; x != nil {
			if err := enrichS3(ctx, &x.S3); err != nil {
				return err
			}
			source.S3 = x
		}
		step.Spec.Sources[i] = source
	}
	return nil
}

func enrichSinks(ctx context.Context) error {
	for i, sink := range step.Spec.Sinks {
		if x := sink.STAN; x != nil {
			if err := enrichSTAN(ctx, x); err != nil {
				return err
			}
			sink.STAN = x
		} else if x := sink.Kafka; x != nil {
			if err := enrichKafka(ctx, x); err != nil {
				return err
			}
			sink.Kafka = x
		} else if x := sink.S3; x != nil {
			if err := enrichS3(ctx, &x.S3); err != nil {
				return err
			}
			sink.S3 = x
		}
		step.Spec.Sinks[i] = sink
	}
	return nil
}

func leadReplica() bool {
	return replica == 0
}
