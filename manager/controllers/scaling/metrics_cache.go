package scaling

import (
	"container/list"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	lru "github.com/hashicorp/golang-lru"
	pmodel "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// check if the Step is deleted after reaching a number of failed times.
	deadKeyCheckThreashold = 100
)

var (
	httpClient   *http.Client
	metricsCache *lru.Cache

	errMetricsEndpointUnavailable = errors.New("metrics endpoint not available")
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

	// counters for unavailable keys
	deadKeys *sync.Map
}

func NewMetricsCacheHandler(client client.Client, workers int) *MetricsCacheHandler {
	return &MetricsCacheHandler{
		client:  client,
		workers: workers,

		stepMap:  make(map[string]*list.Element),
		stepList: list.New(),
		lock:     sync.RWMutex{},

		deadKeys: new(sync.Map),
	}
}

func (m *MetricsCacheHandler) Contains(key string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.stepMap[key]
	return ok
}

func (m *MetricsCacheHandler) Length() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.stepList.Len()
}

func (m *MetricsCacheHandler) StartWatching(key string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.stepMap[key]; !ok {
		m.stepMap[key] = m.stepList.PushBack(key)
	}
	return nil
}

func (m *MetricsCacheHandler) StopWatching(key string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if e, ok := m.stepMap[key]; ok {
		_ = m.stepList.Remove(e)
		delete(m.stepMap, key)
	}
	return nil
}

func (m *MetricsCacheHandler) Start(ctx context.Context) {
	keyCh := make(chan string)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for i := 1; i <= m.workers; i++ {
		go m.pullMetrics(ctx, i, keyCh)
	}

	go func(cctx context.Context) {
		defer runtime.HandleCrash()
		ticker := time.NewTicker(180 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-cctx.Done():
				return
			case <-ticker.C:
				m.deadKeys.Range(func(k, v interface{}) bool {
					times := v.(int)
					if times < deadKeyCheckThreashold {
						return true
					}
					// namespace/name/headless-svc-name
					key := fmt.Sprint(k)
					s := strings.Split(key, "/")
					if err := m.client.Get(cctx, client.ObjectKey{Namespace: s[0], Name: s[1]}, &dfv1.Step{}); err != nil {
						if apierrors.IsNotFound(err) {
							logger.Info("no corresponding step is found", "key", key)
							if err := m.StopWatching(key); err != nil {
								logger.Error(err, "failed to stop watching", "key", key)
								return true
							}
							m.deadKeys.Delete(k)
						} else {
							logger.Error(err, "failed to query step object", "key", key)
						}
						return true
					}
					// Step is existing, remove it from dead keys
					m.deadKeys.Delete(k)
					return true
				})
			}
		}
	}(ctx)

	assign := func() {
		m.lock.Lock()
		defer m.lock.Unlock()
		if m.stepList.Len() == 0 {
			return
		}
		e := m.stepList.Front()
		if key, ok := e.Value.(string); ok {
			m.stepList.MoveToBack(e)
			keyCh <- key
		}
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("exiting metrics caching assigning job")
			return
		default:
			assign()
		}
		time.Sleep(time.Millisecond * time.Duration(func() int {
			l := m.Length()
			if l == 0 {
				return 20000
			}
			result := 20000 / l
			if result > 0 {
				return result
			}
			return 1
		}()))
	}
}

func (m *MetricsCacheHandler) pullMetrics(ctx context.Context, id int, keyCh <-chan string) {
	defer runtime.HandleCrash()
	logger.Info(fmt.Sprintf("started metrics cache worker %v", id))
	for {
		select {
		case <-ctx.Done():
			logger.Info(fmt.Sprintf("stopped metrics cache worker %v", id))
			return
		case key := <-keyCh:
			if pending, err := getPendingMetric(key); err != nil {
				if errors.Is(err, errMetricsEndpointUnavailable) {
					logger.Info("metrics endpoint unavailable, might have been scaled to 0", "key", key)
					if v, existing := m.deadKeys.LoadOrStore(key, 1); existing {
						m.deadKeys.Store(key, v.(int)+1)
					}
				} else {
					logger.Error(err, "failed to get pending messages", "key", key)
				}
			} else {
				pendingKey := key + "/pending"
				lastPendingKey := key + "/last-pending"
				if d, ok := metricsCache.Peek(pendingKey); ok {
					_ = metricsCache.Add(lastPendingKey, d)
				}
				_ = metricsCache.Add(pendingKey, pending)
				// Remove it from dead keys if existing
				m.deadKeys.Delete(key)
			}
		}
	}
}

func getMetrics(key string, replica int) (map[string]*pmodel.MetricFamily, error) {
	// namespace/name/headless-svc-name
	s := strings.Split(key, "/")
	dns := fmt.Sprintf("%s.%s.%s.svc.cluster.local", fmt.Sprintf("%s-%v", s[1], replica), s[2], s[0])
	if _, err := net.LookupIP(dns); err != nil {
		return nil, errMetricsEndpointUnavailable
	}
	endpoint := fmt.Sprintf("https://%s:3570/metrics", dns)
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

func getPendingMetric(key string) (int64, error) {
	metrics, err := getMetrics(key, 0)
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
	if d, ok := metricsCache.Get(fmt.Sprintf("%s/%s/%s/pending", step.Namespace, step.Name, step.GetHeadlessServiceName())); !ok {
		return 0, false
	} else {
		p, yes := d.(int64)
		return p, yes
	}
}

func GetLastPending(step dfv1.Step) (int64, bool) {
	if d, ok := metricsCache.Get(fmt.Sprintf("%s/%s/%s/last-pending", step.Namespace, step.Name, step.GetHeadlessServiceName())); !ok {
		return 0, false
	} else {
		p, yes := d.(int64)
		return p, yes
	}
}
