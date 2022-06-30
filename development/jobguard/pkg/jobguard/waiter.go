package jobguard

import (
	"fmt"
	"log"
	"time"
)

// WaitAtMost executes a function periodically and waits maximum time to finish
func WaitAtMost(fn func() (bool, error), tickTime time.Duration, duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.NewTicker(tickTime)
	defer tick.Stop()

	for {
		if ok, err := fn(); err != nil {
			log.Println(err)
		} else if ok {
			return nil
		}

		select {
		case <-timeout:
			return fmt.Errorf("waiting for resource failed in given timeout %f second(s)", duration.Seconds())
		case <-tick.C:
		}
	}
}
