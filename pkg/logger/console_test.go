package logger

import (
	"bytes"
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newTestConsoleLogger builds a ConsoleLogger that writes to the provided buffers
// instead of os.Stdout/os.Stderr — allows output inspection in tests.
func newTestConsoleLogger(level zapcore.Level, out, errOut *bytes.Buffer) *ConsoleLogger {
	core := &consoleCore{
		level:  level,
		out:    zapcore.AddSync(out),
		errOut: zapcore.AddSync(errOut),
	}
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &ConsoleLogger{SugaredLogger: zapLogger.Sugar()}
}

var _ = Describe("ConsoleLogger", func() {

	// ---------------------------------------------------------------
	// Output routing — Infow → stdout, Errorw → stderr
	// ---------------------------------------------------------------
	Describe("output routing", func() {
		var out, errOut bytes.Buffer
		var logger *ConsoleLogger

		BeforeEach(func() {
			out.Reset()
			errOut.Reset()
			logger = newTestConsoleLogger(zapcore.DebugLevel, &out, &errOut)
		})

		DescribeTable("should route log level to correct output",
			func(logFn func(string, ...interface{}), expectStdout bool) {
				logFn("test message")

				if expectStdout {
					Expect(out.Len()).To(BeNumerically(">", 0), "expected output on stdout")
					Expect(errOut.Len()).To(Equal(0), "expected nothing on stderr")
				} else {
					Expect(errOut.Len()).To(BeNumerically(">", 0), "expected output on stderr")
					Expect(out.Len()).To(Equal(0), "expected nothing on stdout")
				}
			},
			Entry("Debugw → stdout", func(msg string, kv ...interface{}) { logger.Debugw(msg, kv...) }, true),
			Entry("Infow → stdout", func(msg string, kv ...interface{}) { logger.Infow(msg, kv...) }, true),
			Entry("Warnw → stdout", func(msg string, kv ...interface{}) { logger.Warnw(msg, kv...) }, true),
			Entry("Errorw → stderr", func(msg string, kv ...interface{}) { logger.Errorw(msg, kv...) }, false),
		)
	})

	// ---------------------------------------------------------------
	// JSON format — GCP-compatible fields
	// ---------------------------------------------------------------
	Describe("JSON output format", func() {
		var out, errOut bytes.Buffer
		var logger *ConsoleLogger

		BeforeEach(func() {
			out.Reset()
			errOut.Reset()
			logger = newTestConsoleLogger(zapcore.DebugLevel, &out, &errOut)
		})

		It("should include severity, message, and timestamp", func() {
			logger.Infow("hello world")

			var entry map[string]interface{}
			Expect(json.Unmarshal(out.Bytes(), &entry)).To(Succeed())

			Expect(entry).To(HaveKeyWithValue("severity", "INFO"))
			Expect(entry).To(HaveKey("timestamp"))
			Expect(entry).To(HaveKeyWithValue("message", "hello world"))
		})

		DescribeTable("should map levels to GCP severity strings",
			func(logFn func(string, ...interface{}), buf func() *bytes.Buffer, expected string) {
				logFn("msg")
				var entry map[string]interface{}
				Expect(json.Unmarshal(buf().Bytes(), &entry)).To(Succeed())
				Expect(entry).To(HaveKeyWithValue("severity", expected))
			},
			Entry("Debugw → DEBUG", func(msg string, kv ...interface{}) { logger.Debugw(msg, kv...) }, func() *bytes.Buffer { return &out }, "DEBUG"),
			Entry("Infow → INFO", func(msg string, kv ...interface{}) { logger.Infow(msg, kv...) }, func() *bytes.Buffer { return &out }, "INFO"),
			Entry("Warnw → WARNING", func(msg string, kv ...interface{}) { logger.Warnw(msg, kv...) }, func() *bytes.Buffer { return &out }, "WARNING"),
			Entry("Errorw → ERROR", func(msg string, kv ...interface{}) { logger.Errorw(msg, kv...) }, func() *bytes.Buffer { return &errOut }, "ERROR"),
		)

		It("should nest labels under logging.googleapis.com/labels", func() {
			logger.Infow("labeled", LogLabel("app", "my-service"), LogLabel("env", "prod"))

			var entry map[string]interface{}
			Expect(json.Unmarshal(out.Bytes(), &entry)).To(Succeed())

			labels, ok := entry["logging.googleapis.com/labels"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected logging.googleapis.com/labels to be a map")
			Expect(labels).To(HaveKeyWithValue("app", "my-service"))
			Expect(labels).To(HaveKeyWithValue("env", "prod"))
		})

		It("should include sourceLocation", func() {
			logger.Infow("with caller")

			var entry map[string]interface{}
			Expect(json.Unmarshal(out.Bytes(), &entry)).To(Succeed())

			loc, ok := entry["logging.googleapis.com/sourceLocation"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected logging.googleapis.com/sourceLocation to be present")
			Expect(loc).To(HaveKey("file"))
			Expect(loc).To(HaveKey("line"))
			Expect(loc).To(HaveKey("function"))
		})

		It("should include stacktrace on error", func() {
			logger.Errorw("boom")

			var entry map[string]interface{}
			Expect(json.Unmarshal(errOut.Bytes(), &entry)).To(Succeed())

			Expect(entry).To(HaveKey("stacktrace"))
		})

		It("should include structured key-value fields in payload", func() {
			logger.Infow("request", "user_id", "u-123", "status", 200)

			var entry map[string]interface{}
			Expect(json.Unmarshal(out.Bytes(), &entry)).To(Succeed())

			Expect(entry).To(HaveKeyWithValue("user_id", "u-123"))
			Expect(entry).To(HaveKeyWithValue("status", float64(200)))
		})
	})

	// ---------------------------------------------------------------
	// WithSpanContext — trace correlation fields in console JSON
	// ---------------------------------------------------------------
	Describe("WithSpanContext", func() {
		var out, errOut bytes.Buffer
		var logger *ConsoleLogger

		BeforeEach(func() {
			out.Reset()
			errOut.Reset()
			logger = newTestConsoleLogger(zapcore.DebugLevel, &out, &errOut)
		})

		Context("when span context is missing or invalid", func() {

			It("should return ErrInvalidSpanContext for background context", func() {
				child, err := logger.WithSpanContext(context.Background(), "my-project")
				Expect(err).To(MatchError(ErrInvalidSpanContext))
				Expect(child).To(BeNil())
			})

			It("should return ErrInvalidSpanContext for zero span context", func() {
				ctx := trace.ContextWithSpanContext(context.Background(), trace.SpanContext{})
				child, err := logger.WithSpanContext(ctx, "my-project")
				Expect(err).To(MatchError(ErrInvalidSpanContext))
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

			It("should include trace field in console JSON output", func() {
				child, err := logger.WithSpanContext(ctx, "my-gcp-project")
				Expect(err).NotTo(HaveOccurred())

				child.Infow("trace-in-console-test")

				var entry map[string]interface{}
				Expect(json.Unmarshal(out.Bytes(), &entry)).To(Succeed())
				Expect(entry).To(HaveKeyWithValue("message", "trace-in-console-test"), "verifying correct log entry")
				Expect(entry).To(HaveKeyWithValue("logging.googleapis.com/trace",
					"projects/my-gcp-project/traces/4bf92f3577b34da6a3ce929d0e0e4736"))
			})

			It("should include spanId field in console JSON output", func() {
				child, err := logger.WithSpanContext(ctx, "my-gcp-project")
				Expect(err).NotTo(HaveOccurred())

				child.Infow("spanid-in-console-test")

				var entry map[string]interface{}
				Expect(json.Unmarshal(out.Bytes(), &entry)).To(Succeed())
				Expect(entry).To(HaveKeyWithValue("message", "spanid-in-console-test"), "verifying correct log entry")
				Expect(entry).To(HaveKeyWithValue("logging.googleapis.com/spanId", "00f067aa0ba902b7"))
			})

			It("should include trace_sampled=true in console JSON output", func() {
				child, err := logger.WithSpanContext(ctx, "my-gcp-project")
				Expect(err).NotTo(HaveOccurred())

				child.Infow("sampled-true-console-test")

				var entry map[string]interface{}
				Expect(json.Unmarshal(out.Bytes(), &entry)).To(Succeed())
				Expect(entry).To(HaveKeyWithValue("message", "sampled-true-console-test"), "verifying correct log entry")
				Expect(entry).To(HaveKeyWithValue("logging.googleapis.com/trace_sampled", true))
			})

			It("should include all trace fields together in console JSON output", func() {
				child, err := logger.WithSpanContext(ctx, "my-gcp-project")
				Expect(err).NotTo(HaveOccurred())

				child.Infow("all-trace-fields-console-test")

				var entry map[string]interface{}
				Expect(json.Unmarshal(out.Bytes(), &entry)).To(Succeed())
				Expect(entry).To(HaveKeyWithValue("message", "all-trace-fields-console-test"), "verifying correct log entry")
				Expect(entry).To(HaveKeyWithValue("logging.googleapis.com/trace",
					"projects/my-gcp-project/traces/4bf92f3577b34da6a3ce929d0e0e4736"))
				Expect(entry).To(HaveKeyWithValue("logging.googleapis.com/spanId", "00f067aa0ba902b7"))
				Expect(entry).To(HaveKeyWithValue("logging.googleapis.com/trace_sampled", true))
			})

			It("should preserve regular fields alongside trace fields", func() {
				child, err := logger.WithSpanContext(ctx, "my-gcp-project")
				Expect(err).NotTo(HaveOccurred())

				child.Infow("trace-with-fields-test", "user_id", "u-42", "action", "login")

				var entry map[string]interface{}
				Expect(json.Unmarshal(out.Bytes(), &entry)).To(Succeed())
				Expect(entry).To(HaveKeyWithValue("message", "trace-with-fields-test"), "verifying correct log entry")
				Expect(entry).To(HaveKeyWithValue("user_id", "u-42"))
				Expect(entry).To(HaveKeyWithValue("action", "login"))
				Expect(entry).To(HaveKey("logging.googleapis.com/trace"))
			})
		})

		Context("when span context is valid but not sampled", func() {

			It("should include trace_sampled=false in console JSON output", func() {
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

				child.Infow("not-sampled-console-test")

				var entry map[string]interface{}
				Expect(json.Unmarshal(out.Bytes(), &entry)).To(Succeed())
				Expect(entry).To(HaveKeyWithValue("message", "not-sampled-console-test"), "verifying correct log entry")
				Expect(entry).To(HaveKeyWithValue("logging.googleapis.com/trace_sampled", false))
				Expect(entry).To(HaveKey("logging.googleapis.com/trace"))
				Expect(entry).To(HaveKey("logging.googleapis.com/spanId"))
			})
		})
	})
})
