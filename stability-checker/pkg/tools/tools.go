//+build tools

package tools

import (
	// NEEDED FOR AUTO GENERATING MOCKS
	_ "github.com/vektra/mockery/cmd/mockery"
	// HAVE LINT VERSIONED
	_ "golang.org/x/lint"
)
