package http

import (
	"io/ioutil"
	"net/http"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/google/uuid"
)

type httpSource struct {
	ready bool
}

func New(sourceURN, sourceName, authorization string, process source.Process) source.Interface {
	h := &httpSource{true}
	http.HandleFunc("/sources/"+sourceName, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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
		if err := process(
			dfv1.ContextWithMeta(ctx, sourceURN, uuid.New().String(), time.Now()),
			msg,
		); err != nil {
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
