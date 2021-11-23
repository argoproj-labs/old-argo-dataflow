package retry

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sRetry "k8s.io/client-go/util/retry"
)

func retryableErrors(err error) bool {
	return errors.IsTimeout(err) || errors.IsServerTimeout(err) || errors.IsTooManyRequests(err)
}

func WithDefaultRetry(fn func() error) error {
	return WithRetry(k8sRetry.DefaultBackoff, fn)
}

func WithRetry(back wait.Backoff, fn func() error) error {
	return k8sRetry.OnError(back, retryableErrors, fn)
}
