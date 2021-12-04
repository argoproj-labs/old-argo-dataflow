package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type httpSource struct {
	ready bool
}

var Noop = func(context.Context, []byte) error { return nil }

func New(ctx context.Context, secretInterface corev1.SecretInterface, pipelineName, stepName, sourceURN, sourceName string, preProcess func(ctx context.Context, data []byte) error, inbox source.Inbox) (string, source.Interface, error) {
	// we don't want to share this secret
	secret, err := secretInterface.Get(ctx, pipelineName+"-"+stepName, metav1.GetOptions{})
	if err != nil {
		return "", nil, fmt.Errorf("failed to get secret %q: %w", stepName, err)
	}
	authorization := string(secret.Data[fmt.Sprintf("sources.%s.http.authorization", sourceName)])
	h := &httpSource{true}
	http.HandleFunc("/sources/"+sourceName, func(w http.ResponseWriter, r *http.Request) {
		println("ALEX", 0)
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		operationName := fmt.Sprintf("http-source-%s", sourceName)
		var span opentracing.Span
		if err != nil {
			span = opentracing.StartSpan(operationName)
		} else {
			span = opentracing.StartSpan(operationName, ext.RPCServerOption(wireContext))
		}
		defer span.Finish()
		if r.Header.Get("Authorization") != authorization {
			w.WriteHeader(403)
			return
		}
		if !h.ready { // if we are not ready, we cannot serve requests
			w.WriteHeader(503)
			_, _ = w.Write([]byte("not ready"))
			return
		}
		data, err := ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		id := r.Header.Get(dfv1.MetaID)
		if id == "" {
			id = uuid.New().String()
		}
		if err := preProcess(ctx, data); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		result := make(chan error, 1)
		inbox <- &source.Msg{
			Meta: dfv1.Meta{
				Source: sourceURN,
				ID:     id,
				Time:   time.Now().Unix(),
			},
			Data: data,
			Ack: func(context.Context) error {
				result <- nil
				return nil
			},
			Nack: func(ctx context.Context, err error) {
				result <- err
			},
		}
		select {
		case err := <-result:
			if err != nil {
				w.WriteHeader(500)
				_, _ = w.Write([]byte(err.Error()))
			} else {
				w.WriteHeader(204)
			}
		case <-ctx.Done():
			w.WriteHeader(500)
			_, _ = w.Write([]byte(ctx.Err().Error()))
		}
	})
	return authorization, h, nil
}

func (s *httpSource) Close() error {
	s.ready = false
	return nil
}
