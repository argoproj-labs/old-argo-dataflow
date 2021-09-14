package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type httpSink struct {
	sinkName string
	header   http.Header
	client   *http.Client
	url      string
}

func New(ctx context.Context, sinkName string, secretInterface corev1.SecretInterface, x dfv1.HTTPSink) (sink.Interface, error) {
	header := http.Header{}
	for _, h := range x.Headers {
		if h.Value != "" {
			header.Add(h.Name, h.Value)
		} else if h.ValueFrom != nil {
			r := h.ValueFrom.SecretKeyRef
			secret, err := secretInterface.Get(ctx, r.Name, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to get secret %q: %w", r.Name, err)
			}
			header.Add(h.Name, string(secret.Data[r.Key]))
		}
	}
	return httpSink{
		sinkName,
		header,
		&http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: x.InsecureSkipVerify},
			},
		},
		x.URL,
	}, nil
}

func (h httpSink) Sink(ctx context.Context, msg []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("http-sink-%s", h.sinkName))
	defer span.Finish()
	req, err := http.NewRequestWithContext(ctx, "POST", h.url, bytes.NewBuffer(msg))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header = h.header
	req.Header.Add(dfv1.MetaSource.String(), dfv1.GetMetaSource(ctx))
	req.Header.Add(dfv1.MetaID.String(), dfv1.GetMetaID(ctx))
	if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
		return fmt.Errorf("failed to inject tracing headers: %w", err)
	}

	if resp, err := h.client.Do(req); err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	} else {
		defer func() { _ = resp.Body.Close() }()
		_, _ = io.Copy(io.Discard, resp.Body)
		if resp.StatusCode >= 300 {
			return fmt.Errorf("failed to send HTTP request: %q", resp.Status)
		}
	}
	return nil
}
