package sleep

import (
	"time"
)

func Exec(duration string) error {
	v, err := time.ParseDuration(duration)
	if err != nil {
		return err
	}
	time.Sleep(v)
	return nil
}
