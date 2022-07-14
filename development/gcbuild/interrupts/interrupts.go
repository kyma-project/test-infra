package interrupts

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

type Runner interface {
	Run(ctx context.Context) error
}

func Exec(r Runner) error {
	interrupt := make(chan os.Signal, 1)
	kill := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(kill, os.Kill)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-kill:
				fmt.Println("Killing...")

			case <-interrupt:
				fmt.Println("Exiting...")
				cancel()
			case <-ctx.Done():
				return
			}
		}
	}()
	return r.Run(ctx)
}
