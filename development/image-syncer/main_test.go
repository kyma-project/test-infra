package main

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"testing"
)

func Test(t *testing.T) {
	imgRef, _ := name.ParseReference("alpine:3.13.5")
	fmt.Println(imgRef)
	repo := imgRef.Context()
	fmt.Println(repo)
	//get, _ := remote.Image(imgRef)
	//fmt.Printf("%+v\n", get)
	//h, _ := get.Digest()
	//fmt.Println(h)
	repo2, _ := name.NewRepository("eu.gcr.io/kyma-project/external/")
	fmt.Println(repo2)
	fmt.Println()
}
