package retry

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

func Times(t int, fn func() error) error {
	return backoff.Retry(fn, backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Second), uint64(t)))
}
