package volume

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/loadbalanced"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type message struct {
	Path string `json:"path"`
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, pipelineName, stepName, sourceName, sourceURN string, x dfv1.VolumeSource, inbox source.Inbox, leadReplica bool) (source.HasPending, error) {
	logger := sharedutil.NewLogger().WithValues("source", sourceName)
	dir := filepath.Join(dfv1.PathVarRun, "sources", sourceName)
	return loadbalanced.New(ctx, secretInterface, loadbalanced.NewReq{
		Logger:       logger,
		PipelineName: pipelineName,
		StepName:     stepName,
		SourceName:   sourceName,
		SourceURN:    sourceURN,
		LeadReplica:  leadReplica,
		Concurrency:  int(x.Concurrency),
		PollPeriod:   x.PollPeriod.Duration,
		Inbox:        inbox,
		PreProcess: func(ctx context.Context, data []byte) error {
			return nil
		},
		ListItems: func() ([]*source.Msg, error) {
			var msgs []*source.Msg
			err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				// TODO - what is this double-dot business?
				if !d.IsDir() && !strings.Contains(path, "..") {
					// trim dir and also trim leading '/'
					key := strings.TrimPrefix(path, dir)[1:]
					msgs = append(msgs, &source.Msg{
						Meta: dfv1.Meta{
							Source: sourceURN,
							ID:     key,
							Time:   time.Now().Unix(),
						},
						Data: []byte(sharedutil.MustJSON(message{Path: path})),
						Ack: func(context.Context) error {
							if !x.ReadOnly {
								return os.Remove(path)
							}
							return nil
						},
						Nack: source.NoopNack,
					})
				}
				return nil
			})
			return msgs, err
		},
	})
}
