package logger

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
})
