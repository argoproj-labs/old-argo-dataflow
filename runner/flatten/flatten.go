package flatten

import (
	"context"
	"encoding/json"

	"github.com/argoproj-labs/argo-dataflow/sdks/golang"

	"github.com/doublerebel/bellows"
)

func Exec(ctx context.Context) error {
	return golang.StartWithContext(ctx, func(ctx context.Context, msg []byte) ([]byte, error) {
		v := make(map[string]interface{})
		if err := json.Unmarshal(msg, &v); err != nil {
			return nil, err
		}
		if data, err := json.Marshal(bellows.Flatten(v)); err != nil {
			return nil, err
		} else {
			return data, nil
		}
	})
}
