package sidecar

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// https://www.loginradius.com/blog/async/tune-the-go-http-client-for-high-performance/
	httpTransport = http.DefaultTransport.(*http.Transport).Clone()
	httpClient    = &http.Client{
		Transport: httpTransport,
		Timeout:   10 * time.Second,
	}
)

func init() {
	httpTransport.MaxIdleConns = 32
	httpTransport.MaxConnsPerHost = 32
	httpTransport.MaxIdleConnsPerHost = 32
}

func connectIn(ctx context.Context, sink func(context.Context, []byte) error) (func(context.Context, []byte) error, error) {
	inFlight := promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem:   "input",
		Name:        "inflight",
		Help:        "Number of in-flight messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#input_inflight",
		ConstLabels: map[string]string{"replica": strconv.Itoa(replica)},
	})
	messageTimeSeconds := promauto.NewHistogram(prometheus.HistogramOpts{
		Subsystem:   "input",
		Name:        "message_time_seconds",
		Help:        "Message time, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#input_message_time_seconds",
		ConstLabels: map[string]string{"replica": strconv.Itoa(replica)},
	})
	in := step.Spec.GetIn()
	if in == nil {
		logger.Info("no in interface configured")
		return func(context.Context, []byte) error {
			return fmt.Errorf("no in interface configured")
		}, nil
	} else if in.FIFO {
		logger.Info("opened input FIFO")
		fifo, err := os.OpenFile(dfv1.PathFIFOIn, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			return nil, fmt.Errorf("failed to open input FIFO: %w", err)
		}
		addStopHook(func(ctx context.Context) error {
			logger.Info("closing FIFO")
			return fifo.Close()
		})
		return func(ctx context.Context, data []byte) error {
			span, _ := opentracing.StartSpanFromContext(ctx, "fifo")
			defer span.Finish()
			inFlight.Inc()
			defer inFlight.Dec()
			if _, err := fifo.Write(data); err != nil {
				return fmt.Errorf("failed to send to main: %w", err)
			}
			if _, err := fifo.Write([]byte("\n")); err != nil {
				return fmt.Errorf("failed to send to main: %w", err)
			}
			return nil
		}, nil
	} else if in.HTTP != nil {
		logger.Info("HTTP in interface configured")
		if len(step.Spec.Sources) > 0 {
			if err := waitReady(ctx); err != nil {
				return nil, err
			}
		} else {
			logger.Info("not waiting for HTTP to be read, this maybe a generator step and so may never be ready")
		}
		addStopHook(waitUnready)
		return func(ctx context.Context, data []byte) error {
			span, ctx := opentracing.StartSpanFromContext(ctx, "messages")
			defer span.Finish()
			inFlight.Inc()
			defer inFlight.Dec()
			start := time.Now()
			defer func() { messageTimeSeconds.Observe(time.Since(start).Seconds()) }()
			req, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:8080/messages", bytes.NewBuffer(data))
			if err != nil {
				return err
			}
			if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
				return fmt.Errorf("failed to inject tracing headers: %w", err)
			}
			if err := dfv1.MetaInject(ctx, req.Header); err != nil {
				return err
			}
			if resp, err := httpClient.Do(req); err != nil {
				return fmt.Errorf("failed to send to main: %w", err)
			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode >= 300 {
					return fmt.Errorf("failed to send to main: %q %q", resp.Status, body)
				}
				if resp.StatusCode == 201 {
					return sink(ctx, body)
				}
			}
			return nil
		}, nil
	} else {
		return nil, fmt.Errorf("in interface misconfigured")
	}
}

func waitReady(ctx context.Context) error {
	const ipcSockPath = "/var/run/argo-dataflow/main.sock"
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to wait for ready: %w", ctx.Err())
		default:
			if _, err := os.Stat(ipcSockPath); os.Getenv(dfv1.EnvUnixDomainSocket) != "false" && err == nil {
				logger.Info("switching to Unix socket", "path", ipcSockPath)
				dialer := &net.Dialer{}
				httpTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.DialContext(ctx, "unix", ipcSockPath)
				}
			}
			logger.Info("waiting for HTTP in interface to be ready")
			if resp, err := httpClient.Get("http://127.0.0.1:8080/ready"); err == nil && resp.StatusCode < 300 {
				logger.Info("HTTP in interface ready")
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func waitUnready(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to wait for un-ready: %w", ctx.Err())
		default:
			logger.Info("waiting for HTTP in interface to be unready")
			if resp, err := httpClient.Get("http://127.0.0.1:8080/ready"); err != nil || resp.StatusCode >= 300 {
				logger.Info("HTTP in interface unready")
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}
