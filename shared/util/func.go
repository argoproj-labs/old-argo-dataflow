package util

import (
	"reflect"
	"runtime"
	"strings"
)

func GetFuncName(i interface{}) string {
	ptr := runtime.FuncForPC(reflect.ValueOf(i).Pointer())
	parts := strings.SplitN(ptr.Name(), ".", 3)
	return parts[2]
}
