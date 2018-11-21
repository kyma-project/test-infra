package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	cli, err := storage.NewClient(ctx)
	panicOnError(err)
	fmt.Println(cli)

}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
