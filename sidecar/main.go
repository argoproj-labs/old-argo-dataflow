package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/Shopify/sarama"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	stopCh := signals.SetupSignalHandler()

	consumer, err := sarama.NewConsumer([]string{os.Getenv("KAFKA_URL")}, &sarama.Config{})
	if err != nil {
		klog.Fatal(err)
	}

	partition, err := consumer.ConsumePartition(os.Getenv("KAFKA_TOPIC"), 1, sarama.OffsetOldest)
	if err != nil {
		klog.Fatal(err)
	}
	defer func() { _ = partition.Close() }()

	for {
		select {
		case <-stopCh:
			return
		case m := <-partition.Messages():
			v := make(map[string]interface{})
			if err := json.Unmarshal(m.Value, &v); err != nil {
				klog.Error(err)
				continue
			}
			e := cloudevents.NewEvent()
			if err := e.SetData("application/json", v); err != nil {
				klog.Error(err)
				continue
			}
			data, err := json.Marshal(e)
			if err != nil {
				klog.Error(err)
				continue
			}
			resp, err := http.Post("http://localhost:8080", "application/json", bytes.NewBuffer(data))
			if err != nil {
				klog.Error(err)
				continue
			}
			if resp.StatusCode != 200 {
				klog.Error(resp.StatusCode)
			}
		}
	}
}
