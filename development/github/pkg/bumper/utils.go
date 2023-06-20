package bumper

import (
	"fmt"
	"io"
	"strings"
)

// GitCommand is used to pass the various components of the git command which needs to be executed
type GitCommand struct {
	baseCommand string
	args        []string
	workingDir  string
}

func (gc *GitCommand) Call(stdout, stderr io.Writer) error {
	return Call(stdout, stderr, gc.baseCommand, gc.buildCommand()...)
}

func (gc *GitCommand) buildCommand() []string {
	var args []string
	if gc.workingDir != "" {
		args = append(args, "-C", gc.workingDir)
	}
	args = append(args, gc.args...)

	return args
}

func (gc *GitCommand) getCommand() string {
	return fmt.Sprintf("%s %s", gc.baseCommand, strings.Join(gc.buildCommand(), " "))
}

// CensoredWriter is wrapper for io.writer which  will censor secrets using provided censor
type CensoredWriter struct {
	Delegate io.Writer
	Censor   func(content []byte) []byte
}

func (w CensoredWriter) Write(content []byte) (int, error) {
	censored := w.Censor(content)
	return w.Delegate.Write(censored)
}
