package wrappers

import (
	"errors"
	"time"
)

type wrappedFunc func() error

func RetryWrapper(f wrappedFunc, attempts int, sleepStep time.Duration) error {
	var timeToSleep = time.Duration(1 * time.Second)
	var allErrors error

	for i := 0; i < attempts; i++ {
		var err error

		if err = f(); err == nil {
			return nil
		}

		allErrors = errors.Join(allErrors, err)
		time.Sleep(timeToSleep)
		timeToSleep = timeToSleep + sleepStep
	}

	return allErrors
}
