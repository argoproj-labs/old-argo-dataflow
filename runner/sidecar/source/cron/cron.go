package cron

import (
	"context"
	"fmt"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var logger = sharedutil.NewLogger()

type cronSource struct {
	crn *cron.Cron
}

func New(ctx context.Context, sourceName, sourceURN string, x dfv1.Cron, process source.Process) (source.Interface, error) {
	crn := cron.New(
		cron.WithParser(cron.NewParser(cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor)),
		cron.WithChain(cron.Recover(logger)),
	)

	go func() {
		defer runtime.HandleCrash()
		crn.Run()
	}()

	_, err := crn.AddFunc(x.Schedule, func() {
		span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("cron-source-%s", sourceName))
		defer span.Finish()
		msg := []byte(time.Now().Format(x.Layout))
		if err := process(
			dfv1.ContextWithMeta(
				ctx,
				dfv1.Meta{
					Source: sourceURN,
					ID:     uuid.New().String(),
					Time:   time.Now().Unix(),
				},
			),
			msg,
		); err != nil {
			logger.Error(err, "failed to process message")
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to schedule cron %q: %w", x.Schedule, err)
	}
	return cronSource{crn}, nil
}

func (s cronSource) Close() error {
	<-s.crn.Stop().Done()
	return nil
}
