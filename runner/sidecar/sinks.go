package sidecar

import (
	"context"
	"fmt"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/http"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink/log"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/paulbellamy/ratecounter"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
)

func connectSinks(ctx context.Context) (func([]byte) error, error) {
	sinks := map[string]func(msg []byte) error{}
	rateCounters := map[string]*ratecounter.RateCounter{}
	for _, sink := range step.Spec.Sinks {
		logger.Info("connecting sink", "sink", sharedutil.MustJSON(sink))
		sinkName := sink.Name
		if _, exists := sinks[sinkName]; exists {
			return nil, fmt.Errorf("duplicate sink named %q", sinkName)
		}
		rateCounters[sinkName] = ratecounter.NewRateCounter(updateInterval)
		if x := sink.STAN; x != nil {
			if f, err := connectSTANSink(ctx, sinkName, x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = f
			}
		} else if x := sink.Kafka; x != nil {
			sinks[sinkName] = connectKafkaSink(x, sinkName)
		} else if x := sink.Log; x != nil {
			sinks[sinkName] = logsink.New().Sink
		} else if x := sink.HTTP; x != nil {
			if y, err := http.New(ctx, kubernetesInterface, namespace, *x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = y.Sink
			}
		} else {
			return nil, fmt.Errorf("sink misconfigured")
		}
	}

	return func(msg []byte) error {
		for sinkName, f := range sinks {
			counter := rateCounters[sinkName]
			counter.Incr(1)
			withLock(func() {
				step.Status.SinkStatues.IncrTotal(sinkName, replica, printable(msg), rateToResourceQuantity(counter))
			})
			if err := f(msg); err != nil {
				withLock(func() { step.Status.SinkStatues.IncrErrors(sinkName, replica, err) })
				return err
			}
		}
		return nil
	}, nil
}


func connectKafkaSink(x *dfv1.Kafka, sinkName string) func(msg []byte) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:   x.Brokers,
		Dialer:    newKafkaDialer(x),
		Topic:     x.Topic,
		BatchSize: 1,
	})
	addStopHook(func(ctx context.Context) error {
		logger.Info("closing kafka write", "sink", sinkName)
		return writer.Close()
	})
	return func(msg []byte) error {
		return writer.WriteMessages(context.Background(), kafka.Message{Value: msg})
	}
}

func connectSTANSink(ctx context.Context, sinkName string, x *dfv1.STAN) (func(msg []byte) error, error) {
	genClientID := func() string {
		// In a particular situation, the stan connection status is inconsistent between stan server and client,
		// the connection is lost from client side, but the server still thinks it's alive. In this case, use
		// the same client ID to reconnect will fail. To avoid that, add a random number in the client ID string.
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		return fmt.Sprintf("%s-%s-%d-sink-%s-%v", pipelineName, stepName, replica, sinkName, r1.Intn(100))
	}

	var conn *stanConn
	var err error
	clientID := genClientID()
	conn, err = ConnectSTAN(ctx, x, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
	}
	addStopHook(func(ctx context.Context) error {
		logger.Info("closing stan connection", "sink", sinkName)
		if conn != nil && !conn.IsClosed() {
			return conn.Close()
		}
		return nil
	})

	go func() {
		defer runtimeutil.HandleCrash()
		logger.Info("starting stan auto reconnection daemon", "sink", sinkName)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Info("exiting stan auto reconnection daemon", "sink", sinkName)
				return
			case <-ticker.C:
				if conn == nil || conn.IsClosed() {
					logger.Info("stan connection lost, reconnecting...", "sink", sinkName)
					clientID := genClientID()
					conn, err = ConnectSTAN(ctx, x, clientID)
					if err != nil {
						logger.Error(err, "failed to reconnect", "sink", sinkName, "clientID", clientID)
						continue
					}
					logger.Info("reconnected to stan server.", "sink", sinkName, "clientID", clientID)
				}
			}
		}
	}()

	f := func(msg []byte) error { return conn.sc.Publish(x.Subject, msg) }
	return f, nil
}
