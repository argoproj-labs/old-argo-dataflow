package js

import (
	"context"
	"errors"
	"fmt"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharednats "github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/nats"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/nats-io/nats.go"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

type jsSource struct {
	conn *nats.Conn
	sub  *nats.Subscription
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, cluster, namespace, pipelineName, stepName, sourceURN, sourceName string, x dfv1.JetStreamSource, inbox source.Inbox) (source.Interface, error) {
	conn, err := sharednats.ConnectNATS(ctx, secretInterface, x.NATSURL, x.Auth)
	if err != nil {
		return nil, err
	}
	js, err := conn.JetStream()
	if err != nil {
		return nil, err
	}
	queueName := sharedutil.GetSourceUID(cluster, namespace, pipelineName, stepName, sourceName)
	durableName := fmt.Sprintf("%s-%s", queueName, sharedutil.MustHash(x.Subject))
	sub, err := js.QueueSubscribe(x.Subject, queueName, func(msg *nats.Msg) {
		if metadata, err := msg.Metadata(); err != nil {
			logger.Error(err, "failed to get message metadata")
		} else {
			inbox <- &source.Msg{
				Meta: dfv1.Meta{
					Source: sourceURN,
					ID:     fmt.Sprintf("%v-%v", metadata.Sequence.Consumer, metadata.Sequence.Stream),
					Time:   metadata.Timestamp.Unix(),
				},
				Data: msg.Data,
				Ack: func(context.Context) error {
					err := msg.Ack()
					if errors.Is(err, nats.ErrBadSubscription) {
						logger.Info("Jet Stream subscription might have been closed", "source", sourceName, "error", err)
						return nil
					}
					return err
				},
				Nack: source.NoopNack,
			}
		}
	}, nats.ManualAck(), nats.Durable(durableName), nats.DeliverNew())
	if err != nil {
		return nil, err
	}

	return &jsSource{
		conn: conn,
		sub:  sub,
	}, nil
}

func (j jsSource) Close() error {
	logger.Info("closing jetstream source connection")
	if !j.conn.IsClosed() {
		j.conn.Close()
	}
	return nil
}

func (j jsSource) GetPending(ctx context.Context) (uint64, error) {
	if consumerInfo, err := j.sub.ConsumerInfo(); err != nil {
		return 0, fmt.Errorf("failed to get consumer info: %w", err)
	} else {
		return consumerInfo.NumPending, nil
	}
}
