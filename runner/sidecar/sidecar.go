package sidecar

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	tls2 "github.com/argoproj-labs/argo-dataflow/runner/sidecar/tls"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger              = sharedutil.NewLogger()
	cluster             = os.Getenv(dfv1.EnvCluster)
	namespace           = os.Getenv(dfv1.EnvNamespace)
	pod                 = os.Getenv(dfv1.EnvPod)
	pipelineName        = os.Getenv(dfv1.EnvPipelineName)
	ready               = false // we are ready to serve HTTP requests, also updates pod status condition
	kubernetesInterface kubernetes.Interface
	secretInterface     corev1.SecretInterface
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
	kubernetesInterface = kubernetes.NewForConfigOrDie(restConfig)
	secretInterface = kubernetesInterface.CoreV1().Secrets(namespace)

	sharedutil.MustUnJSON(os.Getenv(dfv1.EnvStep), &step)

	logger.Info("step", "cluster", cluster, "step", sharedutil.MustJSON(step))

	if cluster == "" {
		// this must be configured in the controller
		return fmt.Errorf("cluster (%q) was not specified", dfv1.EnvCluster)
	}

	stepName = step.Spec.Name

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

	cfg, err := (&jaegercfg.Configuration{
		Disabled:    true,
		ServiceName: fmt.Sprintf("dataflow-step-%s-%s", pipelineName, stepName),
	}).FromEnv()
	if err != nil {
		return err
	}

	logger.Info("jaeger config", "config", sharedutil.MustJSON(cfg))

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
		jaegercfg.Metrics(metrics.NullFactory),
	)
	if err != nil {
		return err
	}
	defer func() { _ = closer.Close() }()

	opentracing.SetGlobalTracer(tracer)

	if err := enrichSpec(ctx); err != nil {
		return err
	}

	logger.Info("generating self-signed certificate")
	cer, err := tls2.GenerateX509KeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate cert: %w", err)
	}

	logger.Info("sidecar config", "stepName", stepName, "pipelineName", pipelineName, "replica", replica, "updateInterval", updateInterval.String())

	defer logger.Info("done")
	defer stop()
	defer preStop("defer")

	sink, err := connectSinks(ctx)
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
		}, func() float64 {
			if ips, err := net.LookupIP(fmt.Sprintf("%s.%s.svc", step.GetHeadlessServiceName(), namespace)); err != nil {
				return 0
			} else {
				return float64(len(ips))
			}
		})
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

	connectOut(ctx, sink)

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
	httpServer := &http.Server{Addr: ":3570", TLSConfig: &tls.Config{Certificates: []tls.Certificate{*cer}, MinVersion: tls.VersionTLS12}}
	addStopHook(func(ctx context.Context) error {
		logger.Info("closing HTTPS server")
		return httpServer.Shutdown(context.Background())
	})
	go func() {
		defer runtimeutil.HandleCrash()
		logger.Info("starting HTTPS server")
		if err := httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "failed to listen-and-server on HTTPS")
		}
		logger.Info("HTTPS server shutdown")
	}()

	process, err := connectIn(ctx, sink)
	if err != nil {
		return err
	}

	if err := connectSources(ctx, process); err != nil {
		return err
	}

	ready = true
	logger.Info("ready")
	<-ctx.Done()
	return nil
}

func enrichSpec(ctx context.Context) error {
	if err := enrichSources(ctx); err != nil {
		return err
	}
	return enrichSinks(ctx)
}

func enrichSources(ctx context.Context) error {
	for i, source := range step.Spec.Sources {
		if x := source.HTTP; x != nil {
			if x.ServiceName == "" {
				x.ServiceName = pipelineName + "-" + stepName
			}
			source.HTTP = x
		} else if x := source.STAN; x != nil {
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
			if err := enrichKafka(ctx, &x.Kafka); err != nil {
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
