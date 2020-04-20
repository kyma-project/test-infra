// +build tools

package tools

import (
	_ "github.com/vektra/mockery/cmd/mockery"
	_ "k8s.io/code-generator/cmd/deepcopy-gen"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
