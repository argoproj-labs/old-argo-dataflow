package nats

import (
	"context"
	"fmt"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/nats-io/nats.go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = util.NewLogger()

func ConnectNATS(ctx context.Context, secretInterface corev1.SecretInterface, natsURL string, x *dfv1.NATSAuth) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				logger.Error(err, "nats connection lost")
			} else {
				logger.Info("nats disconnected")
			}
		}),
		nats.ReconnectHandler(func(nnc *nats.Conn) {
			logger.Info("reconnected to nats server")
		}),
	}
	authStrategy := authStrategy(x)
	switch authStrategy {
	case dfv1.NATSAuthToken:
		token, err := getAuthToken(ctx, secretInterface, x)
		if err != nil {
			return nil, err
		}
		opts = append(opts, nats.Token(token))
	default:
	}
	logger.Info("nats auth strategy: " + string(authStrategy))
	if nc, err := nats.Connect(natsURL, opts...); err != nil {
		return nil, fmt.Errorf("failed to connect to nats url=%s: %w", natsURL, err)
	} else {
		logger.Info("connected to nats server")
		return nc, nil
	}
}

func getAuthToken(ctx context.Context, secretInterface corev1.SecretInterface, x *dfv1.NATSAuth) (string, error) {
	if x.Token == nil {
		return "", fmt.Errorf("token secret selector is nil")
	}
	secret, err := secretInterface.Get(ctx, x.Token.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if token, ok := secret.Data[x.Token.Key]; ok {
		return string(token), nil
	}
	return "", fmt.Errorf("key %s not found in secret %s", x.Token.Key, x.Token.Name)
}

func authStrategy(x *dfv1.NATSAuth) dfv1.NATSAuthStrategy {
	if x != nil {
		if x.Token != nil {
			return dfv1.NATSAuthToken
		}
	}
	return dfv1.NATSAuthNone
}
