package expand

import (
	"context"
	"encoding/json"

	"github.com/doublerebel/bellows"

	"github.com/argoproj-labs/argo-dataflow/runner/util"
)

func Exec(ctx context.Context) error {
	return util.Do(ctx, func(msg []byte) ([]byte, error) {
		v := make(map[string]interface{})
		if err := json.Unmarshal(msg, &v); err != nil {
			return nil, err
		}
		if data, err := json.Marshal(bellows.Expand(v)); err != nil {
			return nil, err
		} else {
			return data, nil
		}
	})
}
