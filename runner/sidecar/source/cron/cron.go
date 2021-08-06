package cron

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/runtime"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/robfig/cron/v3"
)

var logger = sharedutil.NewLogger()

type cronSource struct {
	crn *cron.Cron
}

func New(x dfv1.Cron, f source.Func) (source.Interface, error) {
	crn := cron.New(
		cron.WithParser(cron.NewParser(cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor)),
		cron.WithChain(cron.Recover(logger)),
	)

	go func() {
		defer runtime.HandleCrash()
		crn.Run()
	}()

	_, err := crn.AddFunc(x.Schedule, func() {
		msg := []byte(time.Now().Format(x.Layout))
		_ = f(context.Background(), msg)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to schedule cron %q: %w", x.Schedule, err)
	}
	return cronSource{crn: crn}, nil
}

func (s cronSource) Close() error {
	<-s.crn.Stop().Done()
	return nil
}
