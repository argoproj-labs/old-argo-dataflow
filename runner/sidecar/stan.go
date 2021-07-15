package sidecar

import (
	"context"
	"fmt"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func subjectiveStan(x *dfv1.STAN) {
	switch x.SubjectPrefix {
	case dfv1.SubjectPrefixNamespaceName:
		x.Subject = fmt.Sprintf("%s.%s", namespace, x.Subject)
	case dfv1.SubjectPrefixNamespacedPipelineName:
		x.Subject = fmt.Sprintf("%s.%s.%s", namespace, pipelineName, x.Subject)
	}
}

func stanFromSecret(s *dfv1.STAN, secret *corev1.Secret) {
	s.NATSURL = dfv1.StringOr(s.NATSURL, string(secret.Data["natsUrl"]))
	s.NATSMonitoringURL = dfv1.StringOr(s.NATSMonitoringURL, string(secret.Data["natsMonitoringUrl"]))
	s.ClusterID = dfv1.StringOr(s.ClusterID, string(secret.Data["clusterId"]))
	s.SubjectPrefix = dfv1.SubjectPrefixOr(s.SubjectPrefix, dfv1.SubjectPrefix(secret.Data["subjectPrefix"]))
	if _, ok := secret.Data["authToken"]; ok {
		s.Auth = &dfv1.STANAuth{
			Token: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Name,
				},
				Key: "authToken",
			},
		}
	}
}

func enrichSTAN(ctx context.Context, secrets v1.SecretInterface, x *dfv1.STAN) error {
	secret, err := secrets.Get(ctx, "dataflow-stan-"+x.Name, metav1.GetOptions{})
	if err != nil {
		if !apierr.IsNotFound(err) {
			return err
		}
	} else {
		stanFromSecret(x, secret)
	}
	subjectiveStan(x)
	return nil
}

func getSTANAuthToken(ctx context.Context, x *dfv1.STAN) (string, error) {
	if x.AuthStrategy() != dfv1.STANAuthToken {
		return "", fmt.Errorf("auth strategy is not token but %s", x.AuthStrategy())
	}
	if x.Auth == nil || x.Auth.Token == nil {
		return "", fmt.Errorf("token secret selector is nil")
	}
	secrets := kubernetesInterface.CoreV1().Secrets(namespace)
	secret, err := secrets.Get(ctx, x.Auth.Token.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if token, ok := secret.Data[x.Auth.Token.Key]; ok {
		return string(token), nil
	}
	return "", fmt.Errorf("key %s not found in secret %s", x.Auth.Token.Key, x.Auth.Token.Name)
}

type stanConn struct {
	nc *nats.Conn
	sc stan.Conn

	natsConnected bool
	stanConnected bool
}

func (nsc *stanConn) Close() error {
	if nsc.sc != nil {
		err := nsc.sc.Close()
		if err != nil {
			return err
		}
	}
	if nsc.nc != nil && nsc.nc.IsConnected() {
		nsc.nc.Close()
	}
	return nil
}

func (nsc *stanConn) IsClosed() bool {
	if nsc.nc == nil || nsc.sc == nil || !nsc.natsConnected || !nsc.stanConnected || nsc.nc.IsClosed() {
		return true
	}
	return false
}

func ConnectSTAN(ctx context.Context, x *dfv1.STAN, clientID string) (*stanConn, error) {
	conn := &stanConn{}
	opts := []nats.Option{
		// Do not reconnect here but handle reconnction outside
		nats.NoReconnect(),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			conn.natsConnected = false
			logger.Error(err, "nats connection lost")
		}),
		nats.ReconnectHandler(func(nnc *nats.Conn) {
			conn.natsConnected = true
			logger.Info("reconnected to nats server")
		}),
	}
	switch x.AuthStrategy() {
	case dfv1.STANAuthToken:
		token, err := getSTANAuthToken(ctx, x)
		if err != nil {
			return nil, err
		}
		opts = append(opts, nats.Token(token))
	default:
	}
	logger.Info("nats auth strategy: " + string(x.AuthStrategy()))
	nc, err := nats.Connect(x.NATSURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats url=%s subject=%s: %w", x.NATSURL, x.Subject, err)
	}
	logger.Info("connected to nats server", "clientID", clientID)
	conn.nc = nc
	conn.natsConnected = true

	sc, err := stan.Connect(x.ClusterID, clientID, stan.NatsConn(nc), stan.Pings(5, 60),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			conn.stanConnected = false
			logger.Error(err, "stan connection lost", "clientID", clientID)
		}))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
	}
	logger.Info("connected to stan server", "clientID", clientID)
	conn.sc = sc
	conn.stanConnected = true
	return conn, nil
}
