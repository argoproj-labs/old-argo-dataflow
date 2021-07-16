package stan

import (
	"context"
	"fmt"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var logger = util.NewLogger()

type Conn struct {
	nc *nats.Conn
	sc stan.Conn

	natsConnected bool
	stanConnected bool
}

func (nsc *Conn) Close() error {
	if nsc.IsClosed() {
		return nil
	}
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

func (nsc *Conn) IsClosed() bool {
	return nsc == nil || nsc.nc == nil || nsc.sc == nil || !nsc.natsConnected || !nsc.stanConnected || nsc.nc.IsClosed()
}

func (nsc *Conn) Publish(subject string, msg []byte) error {
	return nsc.sc.Publish(subject, msg)
}

func (nsc *Conn) QueueSubscribe(subject, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return nsc.sc.QueueSubscribe(subject, qgroup, cb, opts...)
}

func ConnectSTAN(ctx context.Context, kubernetesInterface kubernetes.Interface, namespace string, x dfv1.STAN, clientID string) (*Conn, error) {
	conn := &Conn{}
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
		token, err := getSTANAuthToken(ctx, kubernetesInterface, namespace, x)
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

func getSTANAuthToken(ctx context.Context, kubernetesInterface kubernetes.Interface, namespace string, x dfv1.STAN) (string, error) {
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
