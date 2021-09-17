package golang

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
)

//go:generate sh gen.sh

func Start(handler func(ctx context.Context, msg []byte) ([]byte, error)) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()
	if err := StartWithContext(ctx, handler); err != nil {
		panic(err)
	}
}

func StartWithContext(ctx context.Context, handler func(ctx context.Context, msg []byte) ([]byte, error)) error {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		ctx := MetaExtract(r.Context(), r.Header)
		out, err := func() ([]byte, error) {
			in, err := ioutil.ReadAll(r.Body)
			_ = r.Body.Close()
			if err != nil {
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
	httpServer := &http.Server{Addr: ":8080"}
	go func() {
		defer HandleCrash()
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	udsServer := &http.Server{}
	listener, err := net.Listen("unix", "/var/run/argo-dataflow/main.sock")
	if err != nil {
		return err
	}
	defer func() { _ = listener.Close() }()
	go func() {
		defer HandleCrash()
		if err := udsServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	log.Println("ready")
	defer log.Println("done")
	<-ctx.Done()
	if err := httpServer.Shutdown(context.Background()); err != nil {
		return err
	}
	if err := udsServer.Shutdown(context.Background()); err != nil {
		return err
	}
	return nil
}
