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

func New(ctx context.Context, secretInterface corev1.SecretInterface, pipelineName, stepName, sourceName, sourceURN string, x dfv1.VolumeSource, buffer source.Buffer, leadReplica bool) (source.HasPending, error) {
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
		Process: func(ctx context.Context, msg []byte) error {
			key := string(msg)
			path := filepath.Join(dir, key)
			buffer <- &source.Msg{
				Meta: dfv1.Meta{
					Source: sourceURN,
					ID:     key,
					Time:   time.Now().Unix(),
				},
				Data: []byte(sharedutil.MustJSON(message{Path: path})),
				Ack: func() error {
					if !x.ReadOnly {
						path := filepath.Join(dir, key)
						return os.Remove(path)
					}
					return nil
				},
			}
			return nil
		},
		ListItems: func() ([]interface{}, error) {
			var keys []interface{}
			err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				// TODO - what is this double-dot business?
				if !d.IsDir() && !strings.Contains(path, "..") {
					// trim dir and also trim leading '/'
					keys = append(keys, strings.TrimPrefix(path, dir)[1:])
				}
				return nil
			})
			return keys, err
		},
	})
}
