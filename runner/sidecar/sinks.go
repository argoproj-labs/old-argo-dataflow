package sidecar

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/nats-io/stan.go"
	"github.com/paulbellamy/ratecounter"
)

func connectSinks() (func([]byte) error, error) {
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
			sinks[sinkName] = connectHTTPSink(x)
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

func connectHTTPSink(x *dfv1.HTTPSink) func(msg []byte) error {
	return func(msg []byte) error {
		if resp, err := http.Post(x.URL, "application/octet-stream", bytes.NewBuffer(msg)); err != nil {
			return err
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode >= 300 {
				return fmt.Errorf("failed to send HTTP request: %q %q", resp.Status, body)
			}
		}
		return nil
	}
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
	afterClosers = append(afterClosers, func(ctx context.Context) error {
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
	clientID := fmt.Sprintf("%s-%s-%d-sink-%s", pipelineName, stepName, replica, sinkName)
	sc, err := stan.Connect(x.ClusterID, clientID, stan.NatsURL(x.NATSURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
	}
	afterClosers = append(afterClosers, func(ctx context.Context) error {
		logger.Info("closing stan connection", "sink", sinkName)
		return sc.Close()
	})
	f := func(msg []byte) error { return sc.Publish(x.Subject, msg) }
	return f, nil
}
