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

func New(ctx context.Context, secretInterface corev1.SecretInterface, pipelineName, stepName, sourceURN, sourceName string, process source.Process) (string, source.Interface, error) {
	// we don't want to share this secret
	secret, err := secretInterface.Get(ctx, pipelineName+"-"+stepName, metav1.GetOptions{})
	if err != nil {
		return "", nil, fmt.Errorf("failed to get secret %q: %w", stepName, err)
	}
	authorization := string(secret.Data[fmt.Sprintf("sources.%s.http.authorization", sourceName)])
	h := &httpSource{true}
	http.HandleFunc("/sources/"+sourceName, func(w http.ResponseWriter, r *http.Request) {
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		operationName := fmt.Sprintf("http-source-%s", sourceName)
		var span opentracing.Span
		if err != nil {
			span = opentracing.StartSpan(operationName)
		} else {
			span = opentracing.StartSpan(operationName, ext.RPCServerOption(wireContext))
		}
		defer span.Finish()
		ctx := opentracing.ContextWithSpan(r.Context(), span)
		if r.Header.Get("Authorization") != authorization {
			w.WriteHeader(403)
			return
		}
		if !h.ready { // if we are not ready, we cannot serve requests
			w.WriteHeader(503)
			_, _ = w.Write([]byte("not ready"))
			return
		}
		msg, err := ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		if err := process(
			dfv1.ContextWithMeta(
				ctx,
				dfv1.Meta{
					Source: sourceURN,
					ID:     uuid.New().String(),
					Time:   time.Now().Unix(),
				},
			),
			msg,
		); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(204)
		}
	})
	return authorization, h, nil
}

func (s *httpSource) Close() error {
	s.ready = false
	return nil
}
