package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	go func() {
		select {
		case <-sig:
			fmt.Println("Exiting...")
			os.Exit(0)
		}
	}()
}
