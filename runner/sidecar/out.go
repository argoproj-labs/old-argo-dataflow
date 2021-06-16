package sidecar

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"

	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
)

func connectOut(toSinks func([]byte) error) {
	connectOutFIFO(toSinks)
	connectOutHTTP(toSinks)
}

func connectOutHTTP(f func([]byte) error) {
	logger.Info("HTTP out interface configured")
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+os.Getenv(dfv1.EnvDataflowBearerToken) {
			w.WriteHeader(403)
			return
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error(err, "failed to read message body from main via HTTP")
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		if err := f(data); err != nil {
			logger.Error(err, "failed to send message from main to sink")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(204)
		}
	})
}

func connectOutFIFO(f func([]byte) error) {
	logger.Info("FIFO out interface configured")
	go func() {
		defer runtimeutil.HandleCrash()
		err := func() error {
			fifo, err := os.OpenFile(dfv1.PathFIFOOut, os.O_RDONLY, os.ModeNamedPipe)
			if err != nil {
				return fmt.Errorf("failed to open output FIFO: %w", err)
			}
			afterClosers = append(afterClosers, func(ctx context.Context) error {
				logger.Info("closing out FIFO")
				return fifo.Close()
			})
			logger.Info("opened output FIFO")
			scanner := bufio.NewScanner(fifo)
			for scanner.Scan() {
				if err := f(scanner.Bytes()); err != nil {
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
