package volume

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/loadbalanced"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
)

type message struct {
	Path string `json:"path"`
}

func New(ctx context.Context, pipelineName, stepName, sourceName string, x dfv1.VolumeSource, process source.Process, leadReplica bool) (source.HasPending, error) {
	dir := filepath.Join(dfv1.PathVarRun, "sources", sourceName)
	logger := sharedutil.NewLogger()
	return loadbalanced.New(ctx, loadbalanced.NewReq{
		Logger:       logger,
		PipelineName: pipelineName,
		StepName:     stepName,
		SourceName:   sourceName,
		LeadReplica:  leadReplica,
		Concurrency:  int(x.Concurrency),
		PollPeriod:   x.PollPeriod.Duration,
		Process: func(ctx context.Context, msg []byte) error {
			path := filepath.Join(dir, string(msg))
			return process(ctx, []byte(sharedutil.MustJSON(message{Path: path})))
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
		RemoveItem: func(item interface{}) error {
			if !x.ReadOnly {
				path := filepath.Join(dir, item.(string))
				return os.Remove(path)
			}
			return nil
		},
	})
}
