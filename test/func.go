package test

import (
	"reflect"
	"runtime"
	"strings"
)

func getFuncName(i interface{}) string {
	ptr := runtime.FuncForPC(reflect.ValueOf(i).Pointer())
	parts := strings.SplitN(ptr.Name(), ".", 3)
	return parts[2]
}
