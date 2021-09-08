package sidecar

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

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
		if err := waitReady(ctx); err != nil {
			return nil, err
		}
		addStopHook(waitUnready)
		// https://www.loginradius.com/blog/async/tune-the-go-http-client-for-high-performance/
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConns = 100
		t.MaxConnsPerHost = 100
		t.MaxIdleConnsPerHost = 100
		httpClient := &http.Client{Timeout: 10 * time.Second, Transport: t}
		return func(ctx context.Context, data []byte) error {
			inFlight.Inc()
			defer inFlight.Dec()
			start := time.Now()
			defer func() { messageTimeSeconds.Observe(time.Since(start).Seconds()) }()
			if resp, err := httpClient.Post("http://localhost:8080/messages", "application/octet-stream", bytes.NewBuffer(data)); err != nil {
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
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to wait for ready: %w", ctx.Err())
		default:
			logger.Info("waiting for HTTP in interface to be ready")
			if resp, err := http.Get("http://localhost:8080/ready"); err == nil && resp.StatusCode < 300 {
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
			if resp, err := http.Get("http://localhost:8080/ready"); err != nil || resp.StatusCode >= 300 {
				logger.Info("HTTP in interface unready")
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}
