package loadbalanced

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/go-logr/logr"

	httpsource "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/http"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/workqueue"
)

type loadBalanced struct {
	httpSource source.Interface
	jobs       workqueue.Interface
}

func (s *loadBalanced) GetPending(context.Context) (uint64, error) {
	return uint64(s.jobs.Len()), nil
}

type NewReq struct {
	Logger       logr.Logger
	PipelineName string
	StepName     string
	SourceName   string
	LeadReplica  bool
	Concurrency  int
	PollPeriod   time.Duration
	F            source.Func
	RemoveItem   func(item interface{}) error
	ListItems    func() ([]interface{}, error)
}

func New(ctx context.Context, r NewReq) (source.HasPending, error) {
	logger := r.Logger.WithValues("sourceName", r.SourceName)
	jobs := workqueue.New()
	authorization := sharedutil.RandString()
	httpSource := httpsource.New(r.SourceName, authorization, r.F)
	if r.LeadReplica {
		endpoint := "https://" + r.PipelineName + "-" + r.StepName + "/sources/" + r.SourceName
		logger.Info("starting lead workers", "source", r.SourceName, "endpoint", endpoint)

		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConns = 100
		t.MaxConnsPerHost = 100
		t.MaxIdleConnsPerHost = 100
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient := &http.Client{Timeout: 10 * time.Second, Transport: t}

		// create workers to support Concurrency
		for w := 0; w < r.Concurrency; w++ {
			go func() {
				defer runtime.HandleCrash()
				for {
					item, shutdown := jobs.Get()
					if shutdown {
						return
					}
					func() {
						defer jobs.Done(item)
						itemS := item.(string)
						req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(itemS))
						if err != nil {
							logger.Error(err, "failed to create request", "item", item)
						} else {
							req.Header.Set("Authorization", authorization)
							resp, err := httpClient.Do(req)
							if err != nil {
								logger.Error(err, "failed to process item", "item", item)
							} else {
								body, _ := io.ReadAll(resp.Body)
								_ = resp.Body.Close()
								if resp.StatusCode >= 300 {
									err := fmt.Errorf("%q: %q", resp.Status, body)
									logger.Error(err, "failed to process item", "item", item)
								} else {
									logger.Info("deleting item", "item", item)
									if err := r.RemoveItem(item); err != nil {
										logger.Error(err, "failed to delete item", "item", item)
									}
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
		OUTER:
			for {
				select {
				case <-ctx.Done():
					return
				default:
					endpoint := "https://" + r.PipelineName + "-" + r.StepName + "/ready"
					logger.Info("waiting for HTTP service to be ready", "endpoint", endpoint)
					resp, err := httpClient.Get(endpoint)
					if err == nil && resp.StatusCode < 300 {
						break OUTER
					}
					time.Sleep(3 * time.Second)
				}
			}
			ticker := time.NewTicker(r.PollPeriod)
			defer ticker.Stop()
			f := func() {
				list, err := r.ListItems()
				if err != nil {
					logger.Error(err, "failed to list items")
				} else {
					for _, item := range list {
						jobs.Add(item)
					}
				}
			}
			f()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					f()
				}
			}
		}()
	}
	return &loadBalanced{httpSource, jobs}, nil
}

func (s *loadBalanced) Close() error {
	s.jobs.ShutDown()
	return s.httpSource.Close()
}
