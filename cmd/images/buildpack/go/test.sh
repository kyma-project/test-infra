#!/usr/bin/env bash

set -e
docker run --rm "$IMG" bash -c '
set -e
go version
kubebuilder version
kustomize version
jobguard -help

cat<<EOF > /tmp/main.go
package main
import "fmt"
func main() {
fmt.Println("Hello World!")
}
EOF
go run /tmp/main.go
'