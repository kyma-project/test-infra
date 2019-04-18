package printer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kyma-project/test-infra/stability-checker/internal/log"
	"github.com/kyma-project/test-infra/stability-checker/internal/runner"
)

const (
	invertColor = "\033[7m"
	noColor     = "\033[0m"
)

// LogPrinter prints stability-checker logs
type LogPrinter struct {
	stream           *json.Decoder
	requestedTestIDs map[string]struct{}
	out              io.Writer
}

// New returns new instance of LogPrinter
func New(stream *json.Decoder, ids []string) *LogPrinter {
	var mapped map[string]struct{}
	if len(ids) != 0 {
		mapped = make(map[string]struct{})
		for _, id := range ids {
			mapped[id] = struct{}{}
		}
	}

	return &LogPrinter{
		stream:           stream,
		requestedTestIDs: mapped,
		out:              os.Stdout,
	}
}

// PrintFailedTestOutput prints failed tests outputs.
func (l *LogPrinter) PrintFailedTestOutput() error {
	for {
		var e log.Entry
		switch err := l.stream.Decode(&e); err {
		case nil:
		case io.EOF:
			return nil
		default:
			switch skipErr := l.skipNonJSON(); skipErr {
			case nil:
			case io.EOF:
				return nil
			default:
				return skipErr
			}
			continue
		}

		if l.shouldSkipLogMsg(e) {
			continue
		}

		msg := fmt.Sprintf("%s[%s] Output for test id %q %s\n %s \n", invertColor, e.Log.Time, e.Log.TestRunID, noColor, e.Log.Message)
		if _, err := l.out.Write([]byte(msg)); err != nil {
			return err
		}
	}
}

func (l *LogPrinter) skipNonJSON() error {
	r := l.stream.Buffered()
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	for {
		// read characters until an error or beginning of a new json
		ch, _, err := br.ReadRune()
		if err != nil {
			return err
		}
		if ch == '{' {
			err := br.UnreadRune()
			if err != nil {
				return err
			}
			l.stream = json.NewDecoder(br)
			return nil
		}
	}
}

func (l *LogPrinter) shouldSkipLogMsg(entry log.Entry) bool {
	if entry.Level != "error" {
		return true
	}

	if entry.Log.Type != runner.TestOutputLogType {
		return true
	}

	if l.requestedTestIDs != nil {
		if _, found := l.requestedTestIDs[entry.Log.TestRunID]; !found {
			return true
		}
	}

	return false
}
