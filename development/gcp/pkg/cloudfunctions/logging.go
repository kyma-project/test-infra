package cloudfunctions

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// LogEntry defines a log entry.
type LogEntry struct {
	Message  string `json:"message"`
	Severity string `json:"severity,omitempty"`
	// Trace will be the same for one function call, you can use it for filetering in logs
	Trace  string            `json:"logging.googleapis.com/trace,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
	// Cloud Log Viewer allows filtering and display of this as `jsonPayload.component`.
	Component string `json:"component,omitempty"`
}

// String renders an entry structure to the JSON format expected by Cloud Logging.
func (e LogEntry) String() string {
	if e.Severity == "" {
		e.Severity = "INFO"
	}
	out, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("json.Marshal: %v", err)
	}
	return string(out)
}

func NewLogger() *LogEntry {
	return &LogEntry{}
}
func (e *LogEntry) GenerateTraceValue(projectID, traceFunctionName string) *LogEntry {
	randomInt := rand.Int()
	e.Trace = fmt.Sprintf("projects/%s/traces/%s/%d", projectID, traceFunctionName, randomInt)
	return e
}

func (e *LogEntry) WithLabel(key, value string) *LogEntry {
	if e.Labels == nil {
		e.Labels = make(map[string]string)
	}
	e.Labels[key] = value
	return e
}

func (e *LogEntry) WithTrace(trace string) *LogEntry {
	e.Trace = trace
	return e
}

func (e *LogEntry) WithComponent(component string) *LogEntry {
	e.Component = component
	return e
}

func (e LogEntry) LogCritical(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	e.Severity = "CRITICAL"
	e.Message = message
	fmt.Println(e)
	panic(message)
}

func (e LogEntry) LogError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	e.Severity = "ERROR"
	e.Message = message
	fmt.Println(e)
}

func (e LogEntry) LogWarning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	e.Severity = "WARNING"
	e.Message = message
	fmt.Println(e)
}

func (e LogEntry) LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	e.Severity = "INFO"
	e.Message = message
	fmt.Println(e)
}

func (e LogEntry) LogDebug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	e.Severity = "DEBUG"
	e.Message = message
	fmt.Println(e)
}
