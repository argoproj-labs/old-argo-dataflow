// +build test

package test

import (
	"context"
	"fmt"
	. "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/shared/symbol"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"strings"
	"time"
)

var (
	stepInterface = dynamicInterface.Resource(StepGroupVersionResource).Namespace(namespace)
)

func MessagesPending(s Step) bool {
	return !NothingPending(s)
}

func NothingPending(s Step) bool {
	return s.Status.SourceStatuses.GetPending() == 0
}

func TotalSourceMessagesFunc(f func(int) bool) func(s Step) bool {
	return func(s Step) bool { return f(int(s.Status.SourceStatuses.GetTotal())) }
}

func TotalSourceMessages(n int) func(s Step) bool {
	return TotalSourceMessagesFunc(func(t int) bool { return t == n })
}

func TotalSunkMessagesFunc(f func(int) bool) func(s Step) bool {
	return func(s Step) bool { return f(int(s.Status.SinkStatues.GetTotal())) }
}

func LessThanTotalSunkMessages(n int) func(s Step) bool {
	return TotalSunkMessagesFunc(func(t int) bool { return t < n })
}

func TotalSunkMessages(n int) func(s Step) bool {
	return TotalSunkMessagesFunc(func(t int) bool { return t == n })
}

func WaitForStep(opts ...interface{}) {
	var (
		listOptions = metav1.ListOptions{}
		timeout     = 30 * time.Second
		f           = func(s Step) bool { return s.Status.Phase == StepRunning }
	)
	for _, o := range opts {
		switch v := o.(type) {
		case string:
			listOptions.FieldSelector = "metadata.name=" + v
		case time.Duration:
			timeout = v
		case func(Step) bool:
			f = v
		default:
			panic(fmt.Errorf("un-supported option type %T", v))
		}
	}
	log.Printf("waiting for step %q %q\n", sharedutil.MustJSON(listOptions), sharedutil.GetFuncName(f))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	w, err := stepInterface.Watch(ctx, listOptions)
	if err != nil {
		panic(err)
	}
	defer w.Stop()
	for {
		select {
		case <-ctx.Done():
			panic(fmt.Errorf("failed to wait for step: %w", ctx.Err()))
		case e := <-w.ResultChan():
			un, ok := e.Object.(*unstructured.Unstructured)
			if !ok {
				panic(errors.FromObject(e.Object))
			}
			x := StepFromUnstructured(un)
			y := x.Status
			log.Printf("step %q is %s %q (%s -> %s)\n", x.Name, y.Phase, y.Message, formatSourceStatuses(y.SourceStatuses), formatSourceStatuses(y.SinkStatues))
			if f(x) {
				return
			}
		}
	}
}

func formatSourceStatuses(statuses SourceStatuses) string {
	var sourceText []string
	p := message.NewPrinter(language.English) // adds thousand separator, i.e. "1000000" becomes "1,000,000"
	sym := func(s string, n uint64) string {
		if n > 0 {
			return fmt.Sprintf("️%s%d ", s, n)
		}
		return ""
	}
	for _, s := range statuses {
		for _, m := range s.Metrics {
			rate, _ := m.Rate.AsInt64()
			sourceText = append(sourceText, p.Sprintf("%s%s%s%d %d", sym(symbol.Pending, s.GetPending()), sym(symbol.Error, m.Errors), symbol.Rate,  rate, m.Total))
		}
	}
	return strings.Join(sourceText,",")
}
