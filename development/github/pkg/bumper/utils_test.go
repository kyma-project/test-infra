package bumper

import (
	"reflect"
	"testing"
)

func Test_buildCommand(t *testing.T) {
	tc := []struct {
		name     string
		command  GitCommand
		expected []string
	}{
		{
			name: "simple command without working directory",
			command: GitCommand{
				baseCommand: "git",
				args:        []string{"add", "test.txt"},
			},
			expected: []string{"add", "test.txt"},
		},
		{
			name: "command with working directory",
			command: GitCommand{
				baseCommand: "git",
				args:        []string{"add", "test.txt"},
				workingDir:  "test/directory",
			},
			expected: []string{"-C", "test/directory", "add", "test.txt"},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			command := c.command.buildCommand()

			if !reflect.DeepEqual(command, c.expected) {
				t.Errorf("buildCommand(): got %v, expected %v", command, c.expected)
			}
		})
	}
}

type fakeWriter struct {
	result []byte
}

func (w *fakeWriter) Write(content []byte) (int, error) {
	w.result = append(w.result, content...)
	return len(content), nil
}

func TestWrite(t *testing.T) {
	tc := []struct {
		name     string
		content  []byte
		expected []byte
		censor   func(content []byte) []byte
	}{
		{
			name:     "censor don't change output",
			content:  []byte("some string without secrets"),
			expected: []byte("some string without secrets"),
			censor: func(content []byte) []byte {
				return content
			},
		},
		{
			name:     "censored secret",
			content:  []byte("some secret string"),
			expected: []byte("some XXX string"),
			censor: func(content []byte) []byte {
				return []byte("some XXX string")
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			writer := fakeWriter{}
			cw := CensoredWriter{
				Delegate: &writer,
				Censor:   c.censor,
			}

			written, err := cw.Write(c.content)
			if err != nil {
				t.Errorf("got unexpected error: %s", err)
			}

			if written != len(c.expected) {
				t.Errorf("written bytes doesn't match expected: got %d, expected %d", written, len(c.expected))
			}

			if string(writer.result) != string(c.expected) {
				t.Errorf("expected %s, got %s", string(c.expected), string(writer.result))
			}
		})
	}
}
