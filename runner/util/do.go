package util

import (
	"context"
	"io/ioutil"
	"net/http"

	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"

	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
)

var logger = sharedutil.NewLogger()

func Do(ctx context.Context, fn func(msg []byte) ([]byte, error)) error {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()
		out, err := func() ([]byte, error) {
			if in, err := ioutil.ReadAll(r.Body); err != nil {
				return nil, err
			} else {
				return fn(in)
			}
		}()
		if err != nil {
			logger.Error(err, "failed to marshal message")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		if out != nil {
			w.WriteHeader(201)
			_, _ = w.Write(out)
		} else {
			w.WriteHeader(204)
		}
	})

	server := &http.Server{Addr: ":8080"}

	go func() {
		defer runtimeutil.HandleCrash()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	<-ctx.Done()

	return server.Shutdown(context.Background())
}
