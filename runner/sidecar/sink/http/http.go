package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net/http"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type httpSink struct {
	header http.Header
	client *http.Client
	url    string
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, x dfv1.HTTPSink) (sink.Interface, error) {
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
		url:    x.URL,
		header: header,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (h httpSink) Sink(msg []byte) error {
	req, err := http.NewRequest("POST", h.url, bytes.NewBuffer(msg))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header = h.header
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
