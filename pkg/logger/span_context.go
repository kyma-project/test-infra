package logger

import (
	"context"
	"errors"
	"fmt"

	"github.tools.sap/kyma/neighbors-contracts/go/logging/v2"
	"go.opentelemetry.io/otel/trace"
)

// ErrInvalidSpanContext is returned when the span context in the provided context is not valid.
var ErrInvalidSpanContext = errors.New("span context is not valid")


// withSpanContext is the shared implementation used by all logger types.
// It extracts trace fields from the span context and adds them via With().
func withSpanContext(ctx context.Context, baseLogger logging.LoggerInterface, projectID string) (logging.LoggerInterface, error) {
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	if !spanContext.IsValid() {
		return nil, ErrInvalidSpanContext
	}

	return baseLogger.With(
		"logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", projectID, spanContext.TraceID().String()),
		"logging.googleapis.com/spanId", spanContext.SpanID().String(),
		"logging.googleapis.com/trace_sampled", spanContext.IsSampled(),
	), nil
}
