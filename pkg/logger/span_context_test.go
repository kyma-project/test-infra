package logger

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/trace"
)

var _ = Describe("WithSpanContext", func() {

	var logger *BufferLogger

	BeforeEach(func() {
		logger = NewBufferLogger()
	})

	Context("when span context is missing or invalid", func() {

		It("should return ErrInvalidSpanContext for background context", func() {
			child, err := logger.WithSpanContext(context.Background(), "my-project")
			Expect(err).To(MatchError(ContainSubstring("span context is not valid")))
			Expect(err).To(MatchError(ContainSubstring("projectID=my-project")))
			Expect(child).To(BeNil())
		})

		It("should return ErrInvalidSpanContext for zero span context", func() {
			ctx := trace.ContextWithSpanContext(context.Background(), trace.SpanContext{})
			child, err := logger.WithSpanContext(ctx, "my-project")
			Expect(err).To(MatchError(ContainSubstring("span context is not valid")))
			Expect(err).To(MatchError(ContainSubstring("traceID=00000000000000000000000000000000")))
			Expect(err).To(MatchError(ContainSubstring("spanID=0000000000000000")))
			Expect(child).To(BeNil())
		})
	})

	Context("when span context is valid and sampled", func() {

		var ctx context.Context

		BeforeEach(func() {
			traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
			spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
			spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    traceID,
				SpanID:     spanID,
				TraceFlags: trace.FlagsSampled,
			})
			ctx = trace.ContextWithSpanContext(context.Background(), spanCtx)
		})

		It("should add trace resource path", func() {
			child, err := logger.WithSpanContext(ctx, "my-gcp-project")
			Expect(err).NotTo(HaveOccurred())

			child.Infow("msg")
			Expect(logger.Logs()).To(ContainSubstring(
				`"logging.googleapis.com/trace":"projects/my-gcp-project/traces/4bf92f3577b34da6a3ce929d0e0e4736"`))
		})

		It("should add span ID", func() {
			child, err := logger.WithSpanContext(ctx, "my-gcp-project")
			Expect(err).NotTo(HaveOccurred())

			child.Infow("msg")
			Expect(logger.Logs()).To(ContainSubstring(
				`"logging.googleapis.com/spanId":"00f067aa0ba902b7"`))
		})

		It("should set trace_sampled to true", func() {
			child, err := logger.WithSpanContext(ctx, "my-gcp-project")
			Expect(err).NotTo(HaveOccurred())

			child.Infow("msg")
			Expect(logger.Logs()).To(ContainSubstring(
				`"logging.googleapis.com/trace_sampled":true`))
		})
	})

	Context("when span context is valid but not sampled", func() {

		It("should set trace_sampled to false", func() {
			traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
			spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
			spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    traceID,
				SpanID:     spanID,
				TraceFlags: 0,
			})
			ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

			child, err := logger.WithSpanContext(ctx, "my-gcp-project")
			Expect(err).NotTo(HaveOccurred())

			child.Infow("msg")
			Expect(logger.Logs()).To(ContainSubstring(
				`"logging.googleapis.com/trace_sampled":false`))
		})
	})
})
