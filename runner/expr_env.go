package main

import (
	"encoding/json"
	"fmt"
)

func exprEnv(msg []byte) map[string]interface{} {
	return map[string]interface{}{
		"msg": msg,
		"string": func(v interface{}) string {
			switch w := v.(type) {
			case nil:
				return ""
			case []byte:
				return string(w)
			default:
				return fmt.Sprintf("%v", v)
			}
		},
		"bytes": func(v interface{}) []byte {
			switch w := v.(type) {
			case nil:
				return nil
			case string:
				return []byte(w)
			default:
				return []byte(fmt.Sprintf("%v", v))
			}
		},
		"object": func(v interface{}) map[string]interface{} {
			x := make(map[string]interface{})
			switch w := v.(type) {
			case nil:
				return nil
			case []byte:
				if err := json.Unmarshal(w, &x); err != nil {
					panic(err)
				}
				return x
			case string:
				if err := json.Unmarshal([]byte(w), &x); err != nil {
					panic(err)
				}
				return x
			default:
				panic("unknown type")
			}
		},
		"json": func(v interface{}) []byte {
			x, err := json.Marshal(v)
			if err != nil {
				panic(err)
			}
			return x
		},
	}
}
