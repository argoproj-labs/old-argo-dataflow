package sidecar

import (
	"context"
	"fmt"
	"strings"
	"time"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/Shopify/sarama"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	runnerutil "github.com/argoproj-labs/argo-dataflow/runner/util"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	sarama.Logger = runnerutil.NewSaramaStdLogger(logger)
}

func kafkaFromSecret(k *dfv1.Kafka, secret *corev1.Secret) {
	k.Brokers = dfv1.StringsOr(k.Brokers, strings.Split(string(secret.Data["brokers"]), ","))
	k.Version = dfv1.StringOr(k.Version, string(secret.Data["version"]))
	if _, ok := secret.Data["net.tls"]; ok {
		k.NET = &dfv1.KafkaNET{TLS: &dfv1.TLS{}}
	}
}

func newKafkaConfig(k *dfv1.Kafka) (*sarama.Config, error) {
	x := sarama.NewConfig()
	x.ClientID = dfv1.CtrSidecar
	x.Consumer.MaxProcessingTime = 30 * time.Second
	if k.Version != "" {
		v, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kafka version %q: %w", k.Version, err)
		}
		x.Version = v
	}
	if k.NET != nil {
		if k.NET.TLS != nil {
			x.Net.TLS.Enable = true
		}
	}
	return x, nil
}

func enrichKafka(ctx context.Context, secrets v1.SecretInterface, x *dfv1.Kafka) error {
	secret, err := secrets.Get(ctx, "dataflow-kafka-"+x.Name, metav1.GetOptions{})
	if err != nil {
		if !apierr.IsNotFound(err) {
			return err
		}
	} else {
		kafkaFromSecret(x, secret)
	}
	return nil
}
