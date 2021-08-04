package cloudfunctions

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
)

// Entry defines a log entry.
type LogEntry struct {
	Message  string `json:"message"`
	Severity string `json:"severity,omitempty"`
	// Trace will be the same for one function call, you can use it for filetering in logs
	Trace  string            `json:"logging.googleapis.com/trace,omitempty"`
	Labels map[string]string `json:"logging.googleapis.com/operation,omitempty"`
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
		log.Printf("json.Marshal: %v", err)
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

func (e LogEntry) LogCritical(message string) {
	e.Severity = "CRITICAL"
	e.Message = message
	log.Println(e)
	panic(message)
}

func (e LogEntry) LogError(message string) {
	e.Severity = "ERROR"
	e.Message = message
	log.Println(e)
}

func (e LogEntry) LogInfo(message string) {
	e.Severity = "INFO"
	e.Message = message
	log.Println(e)
}
