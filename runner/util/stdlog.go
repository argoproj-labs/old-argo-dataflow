package util

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/go-logr/logr"
)

func NewSaramaStdLogger(l logr.Logger) sarama.StdLogger {
	return stdLogr{l}
}

type stdLogr struct{ logr.Logger }

func (s stdLogr) Print(v ...interface{})                 { s.Info(fmt.Sprint(v...)) }
func (s stdLogr) Printf(format string, v ...interface{}) { s.Info(fmt.Sprintf(format, v...)) }
func (s stdLogr) Println(v ...interface{})               { s.Info(fmt.Sprint(v...)) }
