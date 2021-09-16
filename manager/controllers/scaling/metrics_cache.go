package scaling

import (
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
	client         client.Client
	cachedContexts *sync.Map
}

func NewMetricsCacheHandler(client client.Client) *MetricsCacheHandler {
	return &MetricsCacheHandler{
		client:         client,
		cachedContexts: &sync.Map{},
	}
}

func (m *MetricsCacheHandler) Contains(step *dfv1.Step) (bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(step)
	if err != nil {
		return false, fmt.Errorf("failed to get key for step object: %w", err)
	}
	_, ok := m.cachedContexts.Load(key)
	return ok, nil
}

func (m *MetricsCacheHandler) EnqueueStep(step *dfv1.Step) error {
	key, err := cache.MetaNamespaceKeyFunc(step)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	_, existing := m.cachedContexts.LoadOrStore(key, cancel)
	if existing {
		cancel()
		return nil
	}
	go func() {
		defer runtime.HandleCrash()
		m.startCachingLoop(ctx, step.DeepCopy())
	}()
	return nil
}

func (m *MetricsCacheHandler) DeleteStep(step *dfv1.Step) error {
	key, err := cache.MetaNamespaceKeyFunc(step)
	if err != nil {
		return err
	}
	if obj, existing := m.cachedContexts.LoadAndDelete(key); !existing {
		return nil
	} else {
		if cancel, ok := obj.(context.CancelFunc); ok {
			cancel()
		}
	}
	return nil
}

func (m *MetricsCacheHandler) startCachingLoop(ctx context.Context, step *dfv1.Step) {
	for {
		select {
		case <-ctx.Done():
			logger.Info("exiting metrics fetching loop...", "namespace", step.Namespace, "step", step.Name)
			return
		default:
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
		time.Sleep(30 * time.Second)
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
