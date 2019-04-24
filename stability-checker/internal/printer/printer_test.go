package printer_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kyma-project/test-infra/stability-checker/internal/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogPrinter(t *testing.T) {
	// given
	input := `
	{"level":"error", "log":{"message":"message-one", "test-run-id":"tid-001", "type":"test-output-other-type", "time":"2017-11-02T16:00:00-04:00"}}
	broken line 
	{"level":"error", "log":{"message":"message-two", "test-run-id":"tid-001", "type":"test-output"}}
	{"level":"info", "log":{"message":"message-three", "test-run-id":"tid-001", "type":"test-output"}}
	{"level":"error", "log":{"message":"message-four", "test-run-id":"tid-002", "type":"test-output"}}
	fail`
	jsonReader := strings.NewReader(input)
	decodeStream := json.NewDecoder(jsonReader)
	var output bytes.Buffer
	logPrinter := printer.NewWithOutput(decodeStream, []string{"tid-001"}, &output)

	// when
	err := logPrinter.PrintFailedTestOutput()
	require.NoError(t, err)

	// then
	log := string(output.Bytes())
	assert.NotContains(t, log, "message-one")
	assert.Contains(t, log, "message-two")
	assert.NotContains(t, log, "message-three")
	assert.NotContains(t, log, "message-four")
	assert.NotContains(t, log, "broken")
}
