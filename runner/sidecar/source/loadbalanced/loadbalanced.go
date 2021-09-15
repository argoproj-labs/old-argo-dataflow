package loadbalanced

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	httpsource "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/http"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/workqueue"
)

type loadBalanced struct {
	httpSource source.Interface
	jobs       workqueue.Interface
}

func (s *loadBalanced) GetPending(context.Context) (uint64, error) {
	// (a) if the jobs have yet to be polled, then this will be zero
	// (b) if polling results in fewer results that the true pending amount (e.g. S3 ListObjectV2 returns max 1000 results)
	// the the pending amount will be capped at that amount
	return uint64(s.jobs.Len()), nil
}

type NewReq struct {
	Logger       logr.Logger
	PipelineName string
	StepName     string
	SourceName   string
	SourceURN    string
	LeadReplica  bool
	Concurrency  int
	PollPeriod   time.Duration
	Process      source.Process
	RemoveItem   func(item interface{}) error
	ListItems    func() ([]interface{}, error)
}

func New(ctx context.Context, r NewReq) (source.HasPending, error) {
	logger := r.Logger.WithValues("sourceName", r.SourceName)
	// (a) in the future we could use a named queue to expose metrics
	// (b) it would be good to limit the size of this work queue and have the `Add
	jobs := workqueue.New()
	authorization := sharedutil.RandString()
	httpSource := httpsource.New(r.SourceURN, r.SourceName, authorization, r.Process)
	if r.LeadReplica {
		endpoint := "https://" + r.PipelineName + "-" + r.StepName + "/sources/" + r.SourceName
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConns = 32
		t.MaxConnsPerHost = 32
		t.MaxIdleConnsPerHost = 32
		t.TLSClientConfig.InsecureSkipVerify = true
		httpClient := &http.Client{Timeout: 10 * time.Second, Transport: t}

		logger.Info("starting lead replica's workers", "source", r.SourceName, "endpoint", endpoint)
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
						req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBufferString(itemS))
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
		logger.Info("starting lead replica's change poller")
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
			poll := func() {
				list, err := r.ListItems()
				if err != nil {
					logger.Error(err, "failed to list items")
				} else {
					for _, item := range list {
						jobs.Add(item)
					}
				}
			}
			logger.Info("executing initial poll")
			poll()
			if r.PollPeriod > 0 {
				logger.Info("starting polling loop", "pollPeriod", r.PollPeriod)
				ticker := time.NewTicker(r.PollPeriod)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						poll()
					}
				}
			} else {
				logger.Info("polling loop disabled", "pollPeriod", r.PollPeriod)
			}
		}()
	}
	return &loadBalanced{httpSource, jobs}, nil
}

func (s *loadBalanced) Close() error {
	s.jobs.ShutDown()
	return s.httpSource.Close()
}
