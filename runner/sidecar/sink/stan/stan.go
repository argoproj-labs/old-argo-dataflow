package stan

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/stan"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/opentracing/opentracing-go"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

type stanSink struct {
	sinkName string
	conn     *stan.Conn
	subject  string
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, namespace, pipelineName, stepName string, replica int, sinkName string, x dfv1.STAN) (sink.Interface, error) {
	genClientID := func() string {
		// In a particular situation, the stan connection status is inconsistent between stan server and client,
		// the connection is lost from client side, but the server still thinks it's alive. In this case, use
		// the same client ID to reconnect will fail. To avoid that, add a random number in the client ID string.
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		return fmt.Sprintf("%s-%s-%s-%d-sink-%s-%v", namespace, pipelineName, stepName, replica, sinkName, r1.Intn(100))
	}

	var conn *stan.Conn
	var err error
	clientID := genClientID()
	conn, err = stan.ConnectSTAN(ctx, secretInterface, x, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to stan url=%s clusterID=%s clientID=%s subject=%s: %w", x.NATSURL, x.ClusterID, clientID, x.Subject, err)
	}
	go func() {
		defer runtimeutil.HandleCrash()
		logger.Info("starting stan auto reconnection daemon", "sink", sinkName)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Info("exiting stan auto reconnection daemon", "sink", sinkName)
				return
			case <-ticker.C:
				if conn == nil || conn.IsClosed() {
					logger.Info("stan connection lost, reconnecting...", "sink", sinkName)
					clientID := genClientID()
					conn, err = stan.ConnectSTAN(ctx, secretInterface, x, clientID)
					if err != nil {
						logger.Info("failed to reconnect, will try again soon", "sink", sinkName, "clientID", clientID, "error", err)
						continue
					}
					logger.Info("reconnected to stan server.", "sink", sinkName, "clientID", clientID)
				}
			}
		}
	}()

	return stanSink{sinkName, conn, x.Subject}, nil
}

func (s stanSink) Sink(ctx context.Context, msg []byte) error {
	span, _ := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("stan-sink-%s", s.sinkName))
	defer span.Finish()
	return s.conn.Publish(s.subject, msg)
}

func (s stanSink) Close() error {
	logger.Info("closing stan sink connection")
	return s.conn.Close()
}
