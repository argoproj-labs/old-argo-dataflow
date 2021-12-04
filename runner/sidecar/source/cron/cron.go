package cron

import (
	"fmt"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var logger = sharedutil.NewLogger()

type cronSource struct {
	crn *cron.Cron
}

func New(sourceURN string, x dfv1.Cron, inbox source.Inbox) (source.Interface, error) {
	crn := cron.New(
		cron.WithParser(cron.NewParser(cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor)),
		cron.WithChain(cron.Recover(logger)),
	)

	go func() {
		defer runtime.HandleCrash()
		crn.Run()
	}()

	_, err := crn.AddFunc(x.Schedule, func() {
		now := time.Now()
		msg := []byte(now.Format(x.Layout))
		inbox <- &source.Msg{
			Meta: dfv1.Meta{
				Source: sourceURN,
				ID:     fmt.Sprint(now.Unix()),
				Time:   now.Unix(),
			},
			Data: msg,
			Ack:  source.NoopAck,
			Nack: source.NoopNack,
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
