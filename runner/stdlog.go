package main

import (
	"fmt"

	"github.com/go-logr/logr"
)

type stdlog struct{ logr.Logger }

func (s stdlog) Print(v ...interface{})                 { s.Info(fmt.Sprint(v...)) }
func (s stdlog) Printf(format string, v ...interface{}) { s.Info(fmt.Sprintf(format, v...)) }
func (s stdlog) Println(v ...interface{})               { s.Info(fmt.Sprint(v...)) }
