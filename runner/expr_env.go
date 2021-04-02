package main

import "fmt"

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
	}
}
