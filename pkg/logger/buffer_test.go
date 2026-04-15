package logger

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BufferLogger", func() {
	// logger is recreated before each test — fresh, empty buffer every time.
	var buf *BufferLogger

	BeforeEach(func() {
		buf = NewBufferLogger()
	})

	// ---------------------------------------------------------------
	// Basic creation
	// ---------------------------------------------------------------
	Context("when creating a new BufferLogger", func() {
		It("should not be nil", func() {
			Expect(buf).NotTo(BeNil())
			Expect(buf.Buffer).NotTo(BeNil())
		})

		It("should start with an empty buffer", func() {
			Expect(buf.Logs()).To(BeEmpty())
		})
	})

	// ---------------------------------------------------------------
	// Structured logging methods (Infow, Warnw, Errorw, Debugw)
	// ---------------------------------------------------------------
	DescribeTable("structured logging methods",
		func(logAction func(msg string, kv ...interface{}), expectedSeverity string) {
			// Call the logging method.
			logAction("test message", "key", "value")

			// Sync to flush any buffered data.
			Expect(buf.Sync()).To(Succeed())

			// Parse the JSON output.
			var entry map[string]interface{}
			err := json.Unmarshal([]byte(buf.Logs()), &entry)
			Expect(err).NotTo(HaveOccurred(), "log output should be valid JSON")

			// Check the message.
			Expect(entry).To(HaveKeyWithValue("message", "test message"))

			// Check the structured key-value pair.
			Expect(entry).To(HaveKeyWithValue("key", "value"))

			// Check severity — zapdriver uses "severity" field with uppercase values.
			Expect(entry).To(HaveKeyWithValue("severity", expectedSeverity))
		},
		// zapdriver severity values: DEBUG, INFO, WARNING, ERROR
		Entry("Infow logs at INFO", func(msg string, kv ...interface{}) { buf.Infow(msg, kv...) }, "INFO"),
		Entry("Warnw logs at WARNING", func(msg string, kv ...interface{}) { buf.Warnw(msg, kv...) }, "WARNING"),
		Entry("Errorw logs at ERROR", func(msg string, kv ...interface{}) { buf.Errorw(msg, kv...) }, "ERROR"),
		Entry("Debugw logs at DEBUG", func(msg string, kv ...interface{}) { buf.Debugw(msg, kv...) }, "DEBUG"),
	)

	// ---------------------------------------------------------------
	// With() — child logger with context
	// ---------------------------------------------------------------
	Context("when using With to create a child logger", func() {
		It("should include context fields in all child log entries", func() {
			child := buf.With("request_id", "abc-123")
			child.Infow("handling request", "endpoint", "/info")

			Expect(buf.Sync()).To(Succeed())

			var entry map[string]interface{}
			err := json.Unmarshal([]byte(buf.Logs()), &entry)
			Expect(err).NotTo(HaveOccurred())

			// Fields from With():
			Expect(entry).To(HaveKeyWithValue("request_id", "abc-123"))
			// Fields from the log call:
			Expect(entry).To(HaveKeyWithValue("endpoint", "/info"))
			// The message:
			Expect(entry).To(HaveKeyWithValue("message", "handling request"))
		})

		It("should return a Logger, not a concrete type", func() {
			// This is the key difference from the old package.
			// With() must return Logger so we maintain abstraction.
			var child Logger = buf.With("key", "val")
			Expect(child).NotTo(BeNil())
		})

		It("should write to the same buffer as the parent", func() {
			child := buf.With("component", "auth")
			child.Infow("child log")
			buf.Infow("parent log")

			Expect(buf.Sync()).To(Succeed())

			// Both entries should appear in the same buffer.
			logs := buf.Logs()
			Expect(logs).To(ContainSubstring("child log"))
			Expect(logs).To(ContainSubstring("parent log"))
		})
	})

	// ---------------------------------------------------------------
	// Sync
	// ---------------------------------------------------------------
	Context("when calling Sync", func() {
		It("should not return an error", func() {
			buf.Infow("some log")
			Expect(buf.Sync()).To(Succeed())
		})
	})

	// ---------------------------------------------------------------
	// LogLabel — GCP labels
	// ---------------------------------------------------------------
	Context("when using LogLabel", func() {
		It("should add labels with labels. prefix in log output", func() {
			buf.Infow("labeled log", LogLabel("app", "test-app"), LogLabel("environment", "dev"))

			Expect(buf.Sync()).To(Succeed())

			var entry map[string]interface{}
			err := json.Unmarshal([]byte(buf.Logs()), &entry)
			Expect(err).NotTo(HaveOccurred())

			// BufferLogger uses zapdriver encoder which outputs flat "labels.key" fields.
			// ConsoleLogger and GCPLogger nest these under "logging.googleapis.com/labels".
			Expect(entry).To(HaveKeyWithValue("labels.app", "test-app"))
			Expect(entry).To(HaveKeyWithValue("labels.environment", "dev"))
		})

		It("should include labels from With on every log entry", func() {
			child := buf.With(LogLabel("app", "my-service"))
			child.Infow("first log")
			child.Infow("second log")

			Expect(buf.Sync()).To(Succeed())

			logs := buf.Logs()
			Expect(logs).To(ContainSubstring("first log"))
			Expect(logs).To(ContainSubstring("second log"))

			// Both entries should contain the label.
			Expect(logs).To(ContainSubstring(`"labels.app":"my-service"`))
		})
	})
})
