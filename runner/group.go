package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/antonmedv/expr"
	"github.com/google/uuid"
	"github.com/juju/fslock"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func withLock(dir string, f func() ([][]byte, error)) ([][]byte, error) {
	mu := fslock.New(fmt.Sprintf("%s.lock", dir))
	if err := mu.Lock(); err != nil {
		return nil, fmt.Errorf("failed to lock %s %w", dir, err)
	}
	defer func() { _ = mu.Unlock() }()
	msgs, err := f()
	if err := mu.Unlock(); err != nil {
		return nil, fmt.Errorf("failed to unlock %s %w", dir, err)
	}
	return msgs, err
}

func Group(ctx context.Context, key string, endOfGroup string) error {
	if err := os.Mkdir(dfv1.PathGroups, 0700); IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create groups dir: %w", err)
	}
	prog, err := expr.Compile(key)
	if err != nil {
		return fmt.Errorf("failed to compile %q: %w", key, err)
	}
	endProg, err := expr.Compile(endOfGroup)
	if err != nil {
		return fmt.Errorf("failed to compile %q: %w", endOfGroup, err)
	}
	return do(ctx, func(msg []byte) ([][]byte, error) {
		res, err := expr.Run(prog, exprEnv(msg))
		if err != nil {
			return nil, fmt.Errorf("failed to run program %q: %w", key, err)
		}
		group, ok := res.(string)
		if !ok {
			return nil, fmt.Errorf("key expression must return a string")
		}
		dir := filepath.Join(dfv1.PathGroups, group)
		if err := os.Mkdir(dir, 0700); IgnoreIsExist(err) != nil {
			return nil, fmt.Errorf("failed to create group sub-dir: %w", err)
		}
		return withLock(dir, func() ([][]byte, error) {
			path := filepath.Join(dir, uuid.New().String())
			if err := ioutil.WriteFile(path, msg, 0600); err != nil {
				return nil, fmt.Errorf("failed to create message file: %w", err)
			}
			res, err = expr.Run(endProg, exprEnv(msg))
			if err != nil {
				return nil, fmt.Errorf("failed to run program %q: %w", endOfGroup, err)
			}
			end, ok := res.(bool)
			if !ok {
				return nil, fmt.Errorf("end-of-group expression %q must return a bool", endOfGroup)
			}
			if !end {
				return nil, nil
			}
			items, err := ioutil.ReadDir(dir)
			if err != nil {
				return nil, fmt.Errorf("failed to read dir %q: %w", endOfGroup, err)
			}
			// return items is creating date order, this is only at accuracy of system clock
			sort.Slice(items, func(i, j int) bool {
				return items[i].ModTime().Before(items[j].ModTime())
			})
			msgs := make([][]byte, len(items))
			for i, f := range items {
				msg, err := ioutil.ReadFile(filepath.Join(dfv1.PathGroups, group, f.Name()))
				if err != nil {
					return nil, fmt.Errorf("failed to read file %q: %w", f.Name(), err)
				}
				msgs[i] = msg
			}
			data, err := json.Marshal(msgs)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal messages: %w", err)
			}
			return [][]byte{data}, os.RemoveAll(dir)
		})
	})
}
