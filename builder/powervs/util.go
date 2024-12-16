package powervs

import (
	"fmt"
	"time"
)

// pollUntil validates if a certain condition is met at defined poll intervals.
// If a timeout is reached, an associated error is returned to the caller.
// condition contains the use-case specific code that returns true when a certain condition is achieved.
func pollUntil(pollInterval, timeOut <-chan time.Time, condition func() (bool, error)) error {
	for {
		select {
		case <-timeOut:
			return fmt.Errorf("timed out while waiting for job to complete")
		case <-pollInterval:
			if done, err := condition(); err != nil {
				return err
			} else if done {
				return nil
			}
		}
	}
}
