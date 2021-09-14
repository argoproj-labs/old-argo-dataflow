package golang

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func Start(handler func(ctx context.Context, msg []byte) ([]byte, error)) {
	ctx := SetupSignalsHandler(context.Background())
	if err := StartWithContext(ctx, handler); err != nil {
		panic(err)
	}
}

func StartWithContext(ctx context.Context, handler func(ctx context.Context, msg []byte) ([]byte, error)) error {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		ctx := dfv1.ContextWithMeta(r.Context(), r.Header.Get(dfv1.MetaSource.String()), r.Header.Get(dfv1.MetaID.String()))
		defer func() { _ = r.Body.Close() }()
		out, err := func() ([]byte, error) {
			if in, err := ioutil.ReadAll(r.Body); err != nil {
				return nil, err
			} else {
				return handler(ctx, in)
			}
		}()
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		} else if out != nil {
			w.WriteHeader(201)
			_, _ = w.Write(out)
		} else {
			w.WriteHeader(204)
		}
	})
	// https://medium.com/honestbee-tw-engineer/gracefully-shutdown-in-go-http-server-5f5e6b83da5a
	server := &http.Server{Addr: ":8080"}

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				println(r)
			}
		}()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	log.Println("ready")
	defer log.Println("done")
	<-ctx.Done()
	return server.Shutdown(context.Background())
}
