package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"

	"github.com/Masterminds/sprig"
)

var _sprig = sprig.GenericFuncMap()

var io = map[string]interface{}{
	"cat": cat,
}

func ExprEnv(ctx context.Context, msg []byte) map[string]interface{} {
	source := dfv1.GetMetaSource(ctx)
	id := dfv1.GetMetaID(ctx)
	return map[string]interface{}{
		// values
		"ctx": map[string]string{
			"source": source,
			"id":     id,
		},
		"msg": msg,
		// funcs
		"bytes":  _bytes,
		"int":    _int,
		"json":   _json,
		"string": _string,
		"object": object,
		"sprig":  _sprig,
		"sha1":   _sha1,
		"io":     io,
	}
}

func _bytes(v interface{}) []byte {
	switch w := v.(type) {
	case nil:
		return nil
	case string:
		return []byte(w)
	default:
		return []byte(fmt.Sprintf("%v", v))
	}
}

func cat(v string) []byte {
	data, err := ioutil.ReadFile(v)
	if err != nil {
		panic(err)
	}
	return data
}

func _int(v interface{}) int {
	switch w := v.(type) {
	case []byte:
		i, err := strconv.Atoi(string(w))
		if err != nil {
			panic(fmt.Errorf("cannot convert %q an int", v))
		}
		return i
	case string:
		i, err := strconv.Atoi(w)
		if err != nil {
			panic(fmt.Errorf("cannot convert %q to int", v))
		}
		return i
	case float64:
		return int(w)
	case int:
		return w
	default:
		panic(fmt.Errorf("cannot convert %q to int", v))
	}
}

func _json(v interface{}) []byte {
	x, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return x
}

func _string(v interface{}) string {
	switch w := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(w)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func object(v interface{}) map[string]interface{} {
	x := make(map[string]interface{})
	switch w := v.(type) {
	case nil:
		return nil
	case []byte:
		if err := json.Unmarshal(w, &x); err != nil {
			panic(fmt.Errorf("cannot convert %q to object: %v", v, err))
		}
		return x
	case string:
		if err := json.Unmarshal([]byte(w), &x); err != nil {
			panic(fmt.Errorf("cannot convert %q to object: %v", v, err))
		}
		return x
	default:
		panic("unknown type")
	}
}
