package jetstream

import (
	"context"
	"fmt"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharednats "github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/nats"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/nats-io/nats.go"
	"github.com/opentracing/opentracing-go"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

type jsSink struct {
	sinkName string
	subject  string
	conn     *nats.Conn
	js       nats.JetStreamContext
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, namespace, pipelineName, stepName string, replica int, sinkName string, x dfv1.JetStreamSink) (sink.Interface, error) {
	conn, err := sharednats.ConnectNATS(ctx, secretInterface, x.NATSURL, x.Auth)
	if err != nil {
		return nil, err
	}
	js, err := conn.JetStream()
	if err != nil {
		return nil, err
	}
	return &jsSink{
		sinkName: sinkName,
		subject:  x.Subject,
		conn:     conn,
		js:       js,
	}, nil
}

func (j jsSink) Sink(ctx context.Context, msg []byte) error {
	span, _ := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("jetstream-sink-%s", j.sinkName))
	defer span.Finish()
	m, err := dfv1.MetaFromContext(ctx)
	if err != nil {
		return err
	}
	if _, err := j.js.Publish(j.subject, msg, nats.MsgId(m.ID)); err != nil {
		return err
	}
	return nil
}

func (j jsSink) Close() error {
	logger.Info("closing jetstream sink connection")
	if !j.conn.IsClosed() {
		j.conn.Close()
	}
	return nil
}
