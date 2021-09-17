package scaling

import (
	"container/list"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	lru "github.com/hashicorp/golang-lru"
	pmodel "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	httpClient   *http.Client
	metricsCache *lru.Cache
)

func init() {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 32
	t.MaxConnsPerHost = 32
	t.MaxIdleConnsPerHost = 32
	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient = &http.Client{Timeout: 1 * time.Second, Transport: t}

	metricsCache, _ = lru.New(1000)
}

type MetricsCacheHandler struct {
	client client.Client

	stepMap  map[string]*list.Element
	stepList *list.List
	lock     sync.RWMutex
	workers  int
}

func NewMetricsCacheHandler(client client.Client, workers int) *MetricsCacheHandler {
	return &MetricsCacheHandler{
		client:  client,
		workers: workers,

		stepMap:  make(map[string]*list.Element),
		stepList: list.New(),
		lock:     sync.RWMutex{},
	}
}

func (m *MetricsCacheHandler) Contains(step *dfv1.Step) (bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(step)
	if err != nil {
		return false, fmt.Errorf("failed to get key for step object: %w", err)
	}
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.stepMap[key]
	return ok, nil
}

func (m *MetricsCacheHandler) EnqueueStep(step *dfv1.Step) error {
	key, err := cache.MetaNamespaceKeyFunc(step)
	if err != nil {
		return err
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.stepMap[key]; !ok {
		m.stepMap[key] = m.stepList.PushBack(step)
	}
	return nil
}

func (m *MetricsCacheHandler) DeleteStep(step *dfv1.Step) error {
	key, err := cache.MetaNamespaceKeyFunc(step)
	if err != nil {
		return err
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	if e, ok := m.stepMap[key]; ok {
		_ = m.stepList.Remove(e)
		delete(m.stepMap, key)
	}
	return nil
}

func (m *MetricsCacheHandler) initializeQueue(ctx context.Context) error {
	steps := &dfv1.StepList{}
	if err := m.client.List(ctx, steps, &client.ListOptions{}); err != nil {
		return fmt.Errorf("failed to list all the step objects: %w", err)
	}
	for _, step := range steps.Items {
		if err := m.EnqueueStep(&step); err != nil {
			return fmt.Errorf("failed to put step %s/%s to metrics cache queue: %w", step.Namespace, step.Name, err)
		}
	}
	return nil
}

func (m *MetricsCacheHandler) Start(ctx context.Context) error {
	if err := m.initializeQueue(ctx); err != nil {
		return err
	}

	stepCh := make(chan *dfv1.Step)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for i := 1; i <= m.workers; i++ {
		go m.pullMetrics(ctx, i, stepCh)
	}

	assign := func() {
		m.lock.Lock()
		defer m.lock.Unlock()
		if m.stepList.Len() == 0 {
			return
		}
		e := m.stepList.Front()
		if step, ok := e.Value.(*dfv1.Step); ok {
			m.stepList.MoveToBack(e)
			stepCh <- step
		}
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("exiting metrics caching job assigning task")
			return nil
		default:
			assign()
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (m *MetricsCacheHandler) pullMetrics(ctx context.Context, id int, stepCh <-chan *dfv1.Step) {
	defer runtime.HandleCrash()
	logger.Info(fmt.Sprintf("started metrics cache worker %v", id))
	for {
		select {
		case <-ctx.Done():
			logger.Info(fmt.Sprintf("stopped metrics cache worker %v", id))
			return
		case step := <-stepCh:
			if pending, err := getPendingMetric(*step); err != nil {
				logger.Error(err, "failed to get pending message", "namespace", step.Namespace, "step", step.Name)
			} else {
				pendingKey := step.Namespace + "|" + step.Name + "|pending"
				lastPendingKey := step.Namespace + "|" + step.Name + "|last-pending"
				if d, ok := metricsCache.Peek(pendingKey); ok {
					_ = metricsCache.Add(lastPendingKey, d)
				}
				_ = metricsCache.Add(pendingKey, pending)
			}
		}
	}
}

func getMetrics(step dfv1.Step, replica int) (map[string]*pmodel.MetricFamily, error) {
	endpoint := fmt.Sprintf("https://%s.%s.%s.svc.cluster.local:3570/metrics", fmt.Sprintf("%s-%v", step.Name, replica), step.GetHeadlessServiceName(), step.Namespace)
	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to access %s, error: %w", endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to access %s, unexpected response code: %v", endpoint, resp.StatusCode)
	}
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse prometheus metrics: %w", err)
	}
	return mf, nil
}

func getPendingMetric(step dfv1.Step) (int64, error) {
	metrics, err := getMetrics(step, 0)
	if err != nil {
		return 0, err
	}
	if f, ok := metrics["sources_pending"]; !ok {
		return 0, nil
	} else {
		result := float64(0)
		for _, m := range f.Metric {
			result += m.GetGauge().GetValue()
		}
		return int64(result), nil
	}
}

func GetPending(step dfv1.Step) (int64, bool) {
	if d, ok := metricsCache.Get(step.Namespace + "|" + step.Name + "|pending"); !ok {
		return 0, false
	} else {
		p, yes := d.(int64)
		return p, yes
	}
}

func GetLastPending(step dfv1.Step) (int64, bool) {
	if d, ok := metricsCache.Get(step.Namespace + "|" + step.Name + "|last-pending"); !ok {
		return 0, false
	} else {
		p, yes := d.(int64)
		return p, yes
	}
}
