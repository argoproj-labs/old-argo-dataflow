package sidecar

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
)

func connectOut(ctx context.Context, sink func(context.Context, []byte) error) {
	connectOutFIFO(ctx, sink)
	connectOutHTTP(sink)
}

func connectOutHTTP(sink func(context.Context, []byte) error) {
	logger.Info("HTTP out interface configured")
	v, err := ioutil.ReadFile(dfv1.PathAuthorization)
	if err != nil {
		panic(fmt.Errorf("failed to read authorization file: %w", err))
	}
	authorization := string(v)
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		span, ctx := opentracing.StartSpanFromContext(r.Context(), "/messages")
		defer span.Finish()
		if r.Header.Get("Authorization") != authorization {
			w.WriteHeader(403)
			return
		}
		data, err := ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
		if err != nil {
			logger.Error(err, "failed to read message body from main via HTTP")
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		if err := sink(
			dfv1.ContextWithMeta(ctx, fmt.Sprintf("urn:dataflow:pod:%s.pod.%s.%s:messages", pod, namespace, cluster), uuid.New().String(), time.Now()),
			data,
		); err != nil {
			logger.Error(err, "failed to send message from main to sink")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(204)
		}
	})
}

func connectOutFIFO(ctx context.Context, sink func(context.Context, []byte) error) {
	logger.Info("FIFO out interface configured")
	go func() {
		defer runtimeutil.HandleCrash()
		err := func() error {
			fifo, err := os.OpenFile(dfv1.PathFIFOOut, os.O_RDONLY, os.ModeNamedPipe)
			if err != nil {
				return fmt.Errorf("failed to open output FIFO: %w", err)
			}
			addStopHook(func(ctx context.Context) error {
				logger.Info("closing out FIFO")
				return fifo.Close()
			})
			logger.Info("opened output FIFO")
			scanner := bufio.NewScanner(fifo)
			for scanner.Scan() {
				if err := sink(
					dfv1.ContextWithMeta(ctx, fmt.Sprintf("urn:dataflow:pod:%s.pod.%s.%s:fifo", pod, namespace, cluster), uuid.New().String(), time.Now()),
					scanner.Bytes(),
				); err != nil {
					return fmt.Errorf("failed to send message from main to sink: %w", err)
				}
			}
			if err = scanner.Err(); err != nil {
				return fmt.Errorf("scanner error: %w", err)
			}
			return nil
		}()
		if err != nil {
			logger.Error(err, "failed to received message from FIFO")
			os.Exit(1)
		}
	}()
}
