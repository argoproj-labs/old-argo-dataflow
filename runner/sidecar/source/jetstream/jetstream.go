package js

import (
	"context"
	"fmt"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharednats "github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/nats"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/argoproj-labs/argo-dataflow/shared/util"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/nats-io/nats.go"
	"github.com/opentracing/opentracing-go"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

type jsSource struct {
	conn              *nats.Conn
	sub               *nats.Subscription
	subject           string
	natsMonitoringURL string
	queueName         string
	durableName       string
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, cluster, namespace, pipelineName, stepName, sourceURN string, replica int, sourceName string, x dfv1.JetStreamSource, process source.Process) (source.Interface, error) {
	conn, err := sharednats.ConnectNATS(ctx, secretInterface, x.NATSURL, x.Auth)
	if err != nil {
		return nil, err
	}
	js, err := conn.JetStream()
	if err != nil {
		return nil, err
	}

	queueName := sharedutil.GetSourceUID(cluster, namespace, pipelineName, stepName, sourceName)
	durableName := fmt.Sprintf("%s-%s", queueName, util.MustHash(x.Subject))
	sub, err := js.QueueSubscribe(x.Subject, queueName, func(msg *nats.Msg) {
		span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("jetstream-source-%s", sourceName))
		defer span.Finish()
		if metadata, err := msg.Metadata(); err != nil {
			logger.Error(err, "failed to get message metadata")
		} else {
			if err := process(
				dfv1.ContextWithMeta(ctx, dfv1.Meta{Source: sourceURN, ID: fmt.Sprint(metadata.Sequence.Stream), Time: metadata.Timestamp.Unix()}),
				msg.Data,
			); err != nil {
				logger.Error(err, "failed to process message")
			} else if err := msg.Ack(); err != nil {
				logger.Error(err, "failed to ack message", "source", sourceName)
			}
		}
	}, nats.ManualAck(), nats.Durable(durableName), nats.DeliverNew(), nats.MaxAckPending(int(x.GetMaxInflight())))
	if err != nil {
		return nil, err
	}

	return &jsSource{
		conn:              conn,
		sub:               sub,
		subject:           x.Subject,
		natsMonitoringURL: x.NATSMonitoringURL,
		queueName:         queueName,
		durableName:       durableName,
	}, nil
}

func (j jsSource) Close() error {
	logger.Info("closing jetstream source connection")
	if j.conn.IsClosed() {
		j.conn.Close()
	}
	return nil
}
