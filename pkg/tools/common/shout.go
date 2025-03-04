package common

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func shoutString(message string, args ...interface{}) string {
	var payload = fmt.Sprintf(message, args...)
	const logEntry = "#### %s"
	return fmt.Sprintf(logEntry, payload)
}

// ShoutFirst makes an emphasized log.Info entry
func ShoutFirst(message string, args ...interface{}) {
	log.Info(shoutString(message, args...))
}

// Shout makes an emphasized log.Info entry with an additional empty line above it
func Shout(message string, args ...interface{}) {
	log.Info("")
	log.Info(shoutString(message, args...))
}
# (2025-03-04)