package image

import (
	"fmt"
	"os"
)

// PrintAndFail prints error message and exits program with given exit code
func PrintAndFail(exitCode int, format string, params ...interface{}) {
	fmt.Printf(format, params...)
	os.Exit(exitCode)
}
