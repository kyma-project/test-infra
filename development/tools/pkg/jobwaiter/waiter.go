package jobwaiter

import (
	"fmt"
	"log"
	"time"
)

// WaitAtMost executes a function periodically and waits maximum time to finish
func WaitAtMost(fn func() (bool, error), tickTime time.Duration, duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.Tick(tickTime)

	for {
		ok, err := fn()
		select {
		case <-timeout:
			return fmt.Errorf("waiting for resource failed in given timeout %f second(s)", duration.Seconds())
		case <-tick:
			if err != nil {
				log.Println(err)
			} else if ok {
				return nil
			}
		}
	}
}
