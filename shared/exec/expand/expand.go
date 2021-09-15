package expand

import (
	"context"
	"encoding/json"

	"github.com/doublerebel/bellows"
)

type Impl struct{}

func New() *Impl { return &Impl{} }

func (i *Impl) Init(ctx context.Context) error { return nil }

func (i *Impl) Exec(ctx context.Context, msg []byte) ([]byte, error) {
	v := make(map[string]interface{})
	if err := json.Unmarshal(msg, &v); err != nil {
		return nil, err
	}
	if data, err := json.Marshal(bellows.Expand(v)); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}
