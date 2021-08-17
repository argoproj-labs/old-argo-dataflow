package volume

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	httpsource "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/http"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/workqueue"
)

var logger = sharedutil.NewLogger()

type volumeSource struct {
	httpSource source.Interface
	jobs       workqueue.Interface
	dir        string
}

func (s *volumeSource) GetPending(context.Context) (uint64, error) {
	dir, err := os.ReadDir(s.dir)
	if err != nil {
		return 0, err
	}
	return uint64(len(dir)), nil
}

type message struct {
	Path string `json:"path"`
}

func New(ctx context.Context, pipelineName, stepName, sourceName string, x dfv1.VolumeSource, f source.Func, leadReplica bool) (source.HasPending, error) {
	dir := filepath.Join(dfv1.PathVarRun, "sources", sourceName)
	jobs := workqueue.New()
	authorization := sharedutil.RandString()
	if leadReplica {
		endpoint := "https://" + pipelineName + "-" + stepName + "/sources/" + sourceName
		logger.Info("starting lead workers", "source", sourceName, "endpoint", endpoint)
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConns = 100
		t.MaxConnsPerHost = 100
		t.MaxIdleConnsPerHost = 100
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient := &http.Client{Timeout: 10 * time.Second, Transport: t}
		// create workers to support concurrency
		for w := 0; w < int(x.Concurrency); w++ {
			go func() {
				defer runtime.HandleCrash()
				for {
					item, shutdown := jobs.Get()
					if shutdown {
						return
					}
					func() {
						defer jobs.Done(item)
						path := item.(string)
						req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(path))
						if err != nil {
							panic(err)
						}
						req.Header.Set("Authorization", authorization)
						resp, err := httpClient.Do(req)
						if err != nil {
							logger.Error(err, "failed to process file", "path", path)
						} else {
							body, _ := io.ReadAll(resp.Body)
							_ = resp.Body.Close()
							if resp.StatusCode >= 300 {
								err := fmt.Errorf("%q: %q", resp.Status, body)
								logger.Error(err, "failed to process file", "path", path)
							} else {
								logger.Info("deleting file", "path", path)
								err := os.Remove(path)
								if err != nil {
									logger.Error(err, "failed to delete file", "path", path)
								}
							}
						}
					}()
				}
			}()
		}
		// create leader Goroutine to poll for new files
		go func() {
			defer runtime.HandleCrash()
			ticker := time.NewTicker(x.PollPeriod.Duration)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					list, err := os.ReadDir(dir)
					if err != nil {
						logger.Error(err, "failed to list volume", "dir", dir)
					} else {
						for _, obj := range list {
							jobs.Add(obj)
						}
					}
				}
			}
		}()
	}
	return &volumeSource{
		httpsource.New(sourceName, authorization, func(ctx context.Context, msg []byte) error {
			return f(ctx, []byte(sharedutil.MustJSON(message{Path: string(msg)})))
		}),
		jobs,
		dir,
	}, nil
}

func (s *volumeSource) Close() error {
	s.jobs.ShutDown()
	return s.httpSource.Close()
}
