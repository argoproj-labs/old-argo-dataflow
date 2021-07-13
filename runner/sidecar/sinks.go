package sidecar

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/paulbellamy/ratecounter"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"time"
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
			if f, err := connectSTANSink(sinkName, x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = f
			}
		} else if x := sink.Kafka; x != nil {
			if f, err := connectKafkaSink(x, sinkName); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = f
			}
		} else if x := sink.Log; x != nil {
			sinks[sinkName] = connectLogSink()
		} else if x := sink.HTTP; x != nil {
			if f, err := connectHTTPSink(ctx, x); err != nil {
				return nil, err
			} else {
				sinks[sinkName] = f
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

func connectHTTPSink(ctx context.Context, x *dfv1.HTTPSink) (func(msg []byte) error, error) {
	header := http.Header{}
	for _, h := range x.Headers {
		if h.Value != "" {
			header.Add(h.Name, h.Value)
		} else if h.ValueFrom != nil {
			r := h.ValueFrom.SecretKeyRef
			secret, err := kubernetesInterface.CoreV1().Secrets(namespace).Get(ctx, r.Name, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to get secret %q: %w", r.Name, err)
			}
			header.Add(h.Name, string(secret.Data[r.Key]))
		}
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	return func(msg []byte) error {
		req, err := http.NewRequest("POST", x.URL, bytes.NewBuffer(msg))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}
		req.Header = header
		if resp, err := client.Do(req); err != nil {
			return fmt.Errorf("failed to send HTTP request: %w", err)
		} else {
			defer func() { _ = resp.Body.Close() }()
			_, _ = io.Copy(io.Discard, resp.Body)
			if resp.StatusCode >= 300 {
				return fmt.Errorf("failed to send HTTP request: %q", resp.Status)
			}
		}
		return nil
	}, nil
}

func connectLogSink() func(msg []byte) error {
	return func(msg []byte) error { //nolint:golint,unparam
		logger.Info(string(msg), "type", "log")
		return nil
	}
}

func connectKafkaSink(x *dfv1.Kafka, sinkName string) (func(msg []byte) error, error) {
	config, err := newKafkaConfig(x)
	if err != nil {
		return nil, err
	}
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(x.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}
	addStopHook(func(ctx context.Context) error {
		logger.Info("closing stan producer", "sink", sinkName)
		return producer.Close()
	})
	f := func(msg []byte) error {
		_, _, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: x.Topic,
			Value: sarama.ByteEncoder(msg),
		})
		return err
	}
	return f, nil
}

func connectSTANSink(sinkName string, x *dfv1.STAN) (func(msg []byte) error, error) {
	opts := []nats.Option{}
	switch x.AuthStrategy() {
	case dfv1.STANAuthToken:
		token, err := getSTANAuthToken(context.Background(), x)
		if err != nil {
			return nil, err
		}
		opts = append(opts, nats.Token(token))
	default:
	}
	logger.Info("nats auth strategy: "+string(x.AuthStrategy()), "sink", sinkName)
	nc, err := nats.Connect(x.NATSURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats url=%s subject=%s: %w", x.NATSURL, x.Subject, err)
	}
	addStopHook(func(ctx context.Context) error {
		logger.Info("closing nats connection", "sink", sinkName)
		if nc != nil && nc.IsConnected() {
			nc.Close()
		}
		return nil
	})
	clientID := fmt.Sprintf("%s-%s-%d-sink-%s", pipelineName, stepName, replica, sinkName)
	sc, err := stan.Connect(x.ClusterID, clientID, stan.NatsConn(nc))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
	}
	addStopHook(func(ctx context.Context) error {
		logger.Info("closing stan connection", "sink", sinkName)
		return sc.Close()
	})
	f := func(msg []byte) error { return sc.Publish(x.Subject, msg) }
	return f, nil
}
