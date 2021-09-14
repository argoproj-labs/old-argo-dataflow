package http

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type httpSource struct {
	ready bool
}

func New(sourceName, authorization string, process source.Process) source.Interface {
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
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		if err := process(ctx, msg); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(204)
		}
	})
	return h
}

func (s *httpSource) Close() error {
	s.ready = false
	return nil
}
