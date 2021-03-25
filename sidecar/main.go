package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Shopify/sarama"
	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/klog/klogr"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var log = klogr.New()

func main() {
	if err := mainE(); err != nil {
		log.Error(err, "failed to run main")
	}
}

func sourceToMain(m ce.Event) (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message to send from source to main container: %w", err)
	}
	resp, err := http.Post("http://localhost:8080/messages", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to sent message from source to main container: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("main container returned non-200 status code: %d", resp.StatusCode)
	}
	return m.ID(), nil
}

type handler struct{}

func (handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		e := ce.NewEvent()
		e.SetID(string(uuid.NewUUID()))
		e.SetType("message.dataflow.argoproj.io")
		e.SetSource("dataflow.argoproj.io")
		_ = e.SetData("application/octet-stream", m.Value)
		if id, err := sourceToMain(e); err != nil {
			log.Error(err, "failed to send message from kafka to main container")
		} else {
			log.WithValues("id", id).Info("message sent from kafka to main container")
			sess.MarkMessage(m, "")
		}
	}
	return nil
}

func mainE() error {
	ctx := signals.SetupSignalHandler()

	deploymentName := os.Getenv("DEPLOYMENT_NAME")
	source := dfv1.NewSource(os.Getenv("SOURCE"))
	sink := dfv1.NewSink(os.Getenv("SINK"))

	log.WithValues("source", source, "sink", sink, "deploymentName", deploymentName).Info("config")

	config := sarama.NewConfig()
	config.ClientID = "dataflow-sidecar"

	nc, err := nats.Connect(
		"eventbus-dataflow-stan-svc",
		nats.Name("Argo Dataflow Sidecar for deployment/"+deploymentName),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to bus: %w", err)
	}
	defer nc.Close()

	var mainToSink func(m *ce.Event) error

	if sink.Bus != nil {
		mainToSink = func(m *ce.Event) error {
			data, _ := m.MarshalJSON()
			return nc.Publish(sink.Bus.Subject, data)
		}
	} else if sink.Kafka != nil {
		producer, err := sarama.NewAsyncProducer([]string{sink.Kafka.URL}, config)
		if err != nil {
			return fmt.Errorf("failed to create kafka producer: %w", err)
		}
		mainToSink = func(m *ce.Event) error {
			producer.Input() <- &sarama.ProducerMessage{
				Topic: sink.Kafka.Topic,
				Value: sarama.StringEncoder(m.Data()),
			}
			return nil
		}
	} else {
		return fmt.Errorf("no sink configured")
	}

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		m := &ce.Event{}
		if err := json.NewDecoder(r.Body).Decode(m); err != nil {
			log.Error(err, "failed to decode message from main container ")
			w.WriteHeader(400)
			return
		}
		if err := mainToSink(m); err != nil {
			log.Error(err, "failed to send message from main container to sink")
			w.WriteHeader(500)
			return
		}
		log.WithValues("id", m.ID).Info("message sent from main container to sink")
		w.WriteHeader(200)
	})
	go func() {
		runtime.HandleCrash(runtime.PanicHandlers...)
		log.Info("starting HTTP server")
		err := http.ListenAndServe(":3569", nil)
		if err != nil {
			log.Error(err, "failed to listen-and-server")
			os.Exit(1)
		}
	}()

	if source.Bus != nil {
		if _, err := nc.QueueSubscribe(source.Bus.Subject, deploymentName, func(m *nats.Msg) {
			e := ce.NewEvent()
			if err := e.UnmarshalJSON(m.Data); err != nil {
				log.Error(err, "failed to marshall message from main container to bus")
			} else if id, err := sourceToMain(e); err != nil {
				log.WithValues("id", id).Error(err, "failed to send message from main container to bus")
			} else {
				log.WithValues("id", id).Info("message sent from main container to bus")
			}
		}); err != nil {
			return fmt.Errorf("failed to subscribe: %w", err)
		}
	} else if source.Kafka != nil {
		group, err := sarama.NewConsumerGroup([]string{source.Kafka.URL}, deploymentName, config)
		if err != nil {
			return fmt.Errorf("failed to create kafka consumer group: %w", err)
		}
		defer func() { _ = group.Close() }()
		if err := group.Consume(ctx, []string{source.Kafka.Topic}, handler{}); err != nil {
			return fmt.Errorf("failed to create kafka consumer: %w", err)
		}
	} else {
		return fmt.Errorf("no source configured")
	}

	log.Info("ready")

	<-ctx.Done()

	return nil
}
