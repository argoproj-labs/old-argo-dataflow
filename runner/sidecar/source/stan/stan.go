package stan

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedstan "github.com/argoproj-labs/argo-dataflow/runner/sidecar/shared/stan"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/nats-io/nats-streaming-server/server"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
	"github.com/opentracing/opentracing-go"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var logger = sharedutil.NewLogger()

type stanSource struct {
	sub               stan.Subscription
	conn              *sharedstan.Conn
	subject           string
	natsMonitoringURL string
	queueName         string
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, cluster, namespace, pipelineName, stepName, sourceURN string, replica int, sourceName string, x dfv1.STAN, process source.Process) (source.Interface, error) {
	genClientID := func() string {
		// In a particular situation, the stan connection status is inconsistent between stan server and client,
		// the connection is lost from client side, but the server still thinks it's alive. In this case, use
		// the same client ID to reconnect will fail. To avoid that, add a random number in the client ID string.
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		return fmt.Sprintf("%s-%s-%s-%d-source-%s-%v", namespace, pipelineName, stepName, replica, sourceName, r1.Intn(100))
	}

	var conn *sharedstan.Conn
	var err error
	clientID := genClientID()
	conn, err = sharedstan.ConnectSTAN(ctx, secretInterface, x, clientID)
	if err != nil {
		return nil, err
	}

	// https://docs.nats.io/developing-with-nats-streaming/queues
	var sub stan.Subscription
	queueName := sharedutil.GetSourceUID(cluster, namespace, pipelineName, stepName, sourceName)
	subFunc := func() (stan.Subscription, error) {
		logger.Info("subscribing to STAN queue", "source", sourceName, "queueName", queueName)
		sub, err := conn.QueueSubscribe(x.Subject, queueName, func(msg *stan.Msg) {
			span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("stan-source-%s", sourceName))
			defer span.Finish()
			if err := process(
				dfv1.ContextWithMeta(ctx, dfv1.Meta{Source: sourceURN, ID: fmt.Sprint(msg.Sequence), Time: msg.Timestamp}),
				msg.Data,
			); err != nil {
				logger.Error(err, "failed to process message")
			} else if err := msg.Ack(); err != nil {
				logger.Error(err, "failed to ack message", "source", sourceName)
			}
		}, stan.DurableName(queueName),
			stan.SetManualAckMode(),
			stan.StartAt(pb.StartPosition_NewOnly),
			stan.AckWait(30*time.Second),
			stan.MaxInflight(x.GetMaxInflight()))
		if err != nil {
			return nil, fmt.Errorf("failed to subscribe: %w", err)
		}
		return sub, nil
	}

	sub, err = subFunc()
	if err != nil {
		return nil, err
	}
	go func() {
		defer runtimeutil.HandleCrash()
		logger.Info("starting stan auto reconnection daemon", "source", sourceName)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Info("exiting stan auto reconnection daemon", "source", sourceName)
				return
			case <-ticker.C:
				if conn == nil || conn.IsClosed() {
					_ = sub.Close()
					logger.Info("stan connection lost, reconnecting...", "source", sourceName)
					clientID := genClientID()
					conn, err = sharedstan.ConnectSTAN(ctx, secretInterface, x, clientID)
					if err != nil {
						logger.Error(err, "failed to reconnect", "source", sourceName, "clientID", clientID)
						continue
					}
					logger.Info("reconnected to stan server.", "source", sourceName, "clientID", clientID)
					if sub, err = subFunc(); err != nil {
						logger.Error(err, "failed to subscribe after reconnection", "source", sourceName, "clientID", clientID)
						// Close the connection to let it retry
						_ = conn.Close()
					}
				}
			}
		}
	}()

	return stanSource{
		conn:              conn,
		sub:               sub,
		subject:           x.Subject,
		natsMonitoringURL: x.NATSMonitoringURL,
		queueName:         queueName,
	}, nil
}

func (s stanSource) Close() error {
	logger.Info("closing stan subscription")
	if err := s.sub.Close(); err != nil {
		return err
	}
	logger.Info("closing stan source connection")
	return s.conn.Close()
}

var httpClient = http.Client{
	Timeout: time.Second * 3,
}

func (s stanSource) GetPending(ctx context.Context) (uint64, error) {
	pendingMessages := func(ctx context.Context, channel, queueNameCombo string) (int64, error) {
		monitoringEndpoint := fmt.Sprintf("%s/streaming/channelsz?channel=%s&subs=1", s.natsMonitoringURL, channel)
		req, err := http.NewRequestWithContext(ctx, "GET", monitoringEndpoint, nil)
		if err != nil {
			return 0, err
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return 0, err
		}
		if resp.StatusCode != 200 {
			return 0, fmt.Errorf("invalid response: %s", resp.Status)
		}
		defer func() { _ = resp.Body.Close() }()
		o := server.Channelz{}
		if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
			return 0, err
		}
		maxLastSent := uint64(0)
		for _, s := range o.Subscriptions {
			if s.QueueName == queueNameCombo && s.LastSent > maxLastSent {
				maxLastSent = s.LastSent
			}
		}
		return int64(o.LastSeq - maxLastSent), nil
	}

	// queueNameCombo := {durableName}:{queueGroup}
	queueNameCombo := s.queueName + ":" + s.queueName
	if pending, err := pendingMessages(ctx, s.subject, queueNameCombo); err != nil {
		return 0, fmt.Errorf("failed to get STAN pending for: %w", err)
	} else if pending >= 0 {
		logger.Info("setting STAN pending", "pending", pending)
		return uint64(pending), nil
	}
	return 0, nil
}
