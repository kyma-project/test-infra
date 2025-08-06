package logging

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BufferLogger", func() {
	var (
		logger *BufferLogger
	)

	BeforeEach(func() {
		logger = NewBufferLogger()
	})

	Context("when creating a new BufferLogger", func() {
		It("should not be nil and should have an initialized buffer", func() {
			Expect(logger).NotTo(BeNil())
			Expect(logger.Buffer).NotTo(BeNil())
		})
	})

	DescribeTable("when using SimpleLoggerInterface methods",
		func(logAction func(args ...interface{}), expectedMsg string) {
			logAction(expectedMsg)
			Expect(logger.Logs()).To(ContainSubstring(`"msg":"` + expectedMsg + `"`))
		},
		Entry("should log Info messages", func(args ...interface{}) { logger.Info(args...) }, "info message"),
		Entry("should log Warn messages", func(args ...interface{}) { logger.Warn(args...) }, "warn message"),
		Entry("should log Error messages", func(args ...interface{}) { logger.Error(args...) }, "error message"),
		Entry("should log Debug messages", func(args ...interface{}) { logger.Debug(args...) }, "debug message"),
	)

	DescribeTable("when using FormatedLoggerInterface methods",
		func(logAction func(template string, args ...interface{}), template, expectedMsg string, args ...interface{}) {
			logAction(template, args...)
			Expect(logger.Logs()).To(ContainSubstring(`"msg":"` + expectedMsg + `"`))
		},
		Entry("should log Infof messages", func(template string, args ...interface{}) { logger.Infof(template, args...) }, "info with number %d", "info with number 123", 123),
		Entry("should log Warnf messages", func(template string, args ...interface{}) { logger.Warnf(template, args...) }, "warn with string %s", "warn with string test", "test"),
		Entry("should log Errorf messages", func(template string, args ...interface{}) { logger.Errorf(template, args...) }, "error with bool %t", "error with bool true", true),
		Entry("should log Debugf messages", func(template string, args ...interface{}) { logger.Debugf(template, args...) }, "debug with value %.2f", "debug with value 3.14", 3.14159),
	)

	DescribeTable("when using StructuredLoggerInterface methods",
		func(logAction func(msg string, keysAndValues ...interface{}), msg string, keysAndValues ...interface{}) {
			logAction(msg, keysAndValues...)
			logs := logger.Logs()
			var logData map[string]interface{}
			err := json.Unmarshal([]byte(logs), &logData)

			Expect(err).NotTo(HaveOccurred(), "Log output should be valid JSON")
			Expect(logData).To(HaveKeyWithValue("msg", msg))

			for i := 0; i < len(keysAndValues); i += 2 {
				key := keysAndValues[i].(string)
				value := keysAndValues[i+1]
				Expect(logData).To(HaveKeyWithValue(key, value))
			}
		},
		Entry("should log Infow messages", func(msg string, kv ...interface{}) { logger.Infow(msg, kv...) }, "structured info", "key", "value", "number", 42.0),
		Entry("should log Warnw messages", func(msg string, kv ...interface{}) { logger.Warnw(msg, kv...) }, "structured warn", "status", "pending"),
		Entry("should log Errorw messages", func(msg string, kv ...interface{}) { logger.Errorw(msg, kv...) }, "structured error", "code", 500.0),
		Entry("should log Debugw messages", func(msg string, kv ...interface{}) { logger.Debugw(msg, kv...) }, "structured debug", "user_id", "user-123"),
	)

	Context("when using With method", func() {
		It("should add context to subsequent logs", func() {
			contextLogger := logger.With("request_id", "abc-xyz-123")
			contextLogger.Infow("processing request", "user", "testuser")

			logs := logger.Logs()

			Expect(logs).To(ContainSubstring(`"msg":"processing request"`))
			Expect(logs).To(ContainSubstring(`"user":"testuser"`))
			Expect(logs).To(ContainSubstring(`"request_id":"abc-xyz-123"`))
		})
	})

	Context("when using utility methods", func() {
		It("Logs() should return the content of the buffer", func() {
			logger.Info("first line")
			logger.Info("second line")
			Expect(logger.Logs()).To(And(
				ContainSubstring("first line"),
				ContainSubstring("second line"),
			))
		})

		It("Sync() should not return an error", func() {
			err := logger.Sync()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
