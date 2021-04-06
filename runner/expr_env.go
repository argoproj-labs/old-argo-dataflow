package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Masterminds/sprig"
)

func exprEnv(msg []byte) map[string]interface{} {
	return map[string]interface{}{
		// values
		"msg": msg,
		// funcs
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
		"int": func(v interface{}) int {
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
		},
		"json": func(v interface{}) []byte {
			x, err := json.Marshal(v)
			if err != nil {
				panic(err)
			}
			return x
		},
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
		"object": func(v interface{}) map[string]interface{} {
			x := make(map[string]interface{})
			switch w := v.(type) {
			case nil:
				return nil
			case []byte:
				if err := json.Unmarshal(w, &x); err != nil {
					panic(fmt.Errorf("cannot convert %q to JSON", v))
				}
				return x
			case string:
				if err := json.Unmarshal([]byte(w), &x); err != nil {
					panic(fmt.Errorf("cannot convert %q to JSON", v))
				}
				return x
			default:
				panic("unknown type")
			}
		},
		"sprig": sprig.GenericFuncMap(),
	}
}
