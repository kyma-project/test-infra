package logger

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Factory", func() {

	// ---------------------------------------------------------------
	// parseLogLevel
	// ---------------------------------------------------------------
	Describe("parseLogLevel", func() {
		// After each test, clean up the env var so tests don't affect each other.
		AfterEach(func() {
			os.Unsetenv(EnvLogLevel)
		})

		It("should default to info when LOG_LEVEL is not set", func() {
			os.Unsetenv(EnvLogLevel)
			level, err := parseLogLevel()
			Expect(err).NotTo(HaveOccurred())
			// zapcore.InfoLevel == -1... but we can check the string
			Expect(level.String()).To(Equal("info"))
		})

		It("should return debug level for LOG_LEVEL=debug", func() {
			os.Setenv(EnvLogLevel, "debug")
			level, err := parseLogLevel()
			Expect(err).NotTo(HaveOccurred())
			Expect(level.String()).To(Equal("debug"))
		})

		It("should return info level for LOG_LEVEL=info", func() {
			os.Setenv(EnvLogLevel, "info")
			level, err := parseLogLevel()
			Expect(err).NotTo(HaveOccurred())
			Expect(level.String()).To(Equal("info"))
		})

		It("should handle uppercase and spaces", func() {
			os.Setenv(EnvLogLevel, "  DEBUG  ")
			level, err := parseLogLevel()
			Expect(err).NotTo(HaveOccurred())
			Expect(level.String()).To(Equal("debug"))
		})

		DescribeTable("should return correct level",
			func(input string, expected string) {
				os.Setenv(EnvLogLevel, input)
				level, err := parseLogLevel()
				Expect(err).NotTo(HaveOccurred())
				Expect(level.String()).To(Equal(expected))
			},
			Entry("warn", "warn", "warn"),
			Entry("error", "error", "error"),
			Entry("dpanic", "dpanic", "dpanic"),
			Entry("panic", "panic", "panic"),
			Entry("fatal", "fatal", "fatal"),
		)

		It("should return error for invalid values", func() {
			os.Setenv(EnvLogLevel, "trace")
			_, err := parseLogLevel()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("trace"))
		})
	})

	// ---------------------------------------------------------------
	// parseDestination
	// ---------------------------------------------------------------
	Describe("parseDestination", func() {
		AfterEach(func() {
			os.Unsetenv(EnvLogDestination)
		})

		It("should default to console when LOG_DESTINATION is not set", func() {
			os.Unsetenv(EnvLogDestination)
			dest, err := parseDestination()
			Expect(err).NotTo(HaveOccurred())
			Expect(dest).To(Equal("console"))
		})

		It("should return console for LOG_DESTINATION=console", func() {
			os.Setenv(EnvLogDestination, "console")
			dest, err := parseDestination()
			Expect(err).NotTo(HaveOccurred())
			Expect(dest).To(Equal("console"))
		})

		It("should return api for LOG_DESTINATION=api", func() {
			os.Setenv(EnvLogDestination, "api")
			dest, err := parseDestination()
			Expect(err).NotTo(HaveOccurred())
			Expect(dest).To(Equal("api"))
		})

		It("should return console-and-api for LOG_DESTINATION=console-and-api", func() {
			os.Setenv(EnvLogDestination, "console-and-api")
			dest, err := parseDestination()
			Expect(err).NotTo(HaveOccurred())
			Expect(dest).To(Equal("console-and-api"))
		})

		It("should return error for invalid values", func() {
			os.Setenv(EnvLogDestination, "kafka")
			_, err := parseDestination()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("kafka"))
		})

		It("should return error for old 'both' value", func() {
			os.Setenv(EnvLogDestination, "both")
			_, err := parseDestination()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("both"))
		})
	})

	// ---------------------------------------------------------------
	// New (full factory)
	// ---------------------------------------------------------------
	Describe("New", func() {
		AfterEach(func() {
			os.Unsetenv(EnvLogDestination)
			os.Unsetenv(EnvLogLevel)
		})

		It("should create a ConsoleLogger when LOG_DESTINATION=console", func() {
			os.Setenv(EnvLogDestination, "console")
			logger, err := New()
			Expect(err).NotTo(HaveOccurred())
			Expect(logger).NotTo(BeNil())

			// Verify it's a ConsoleLogger under the hood.
			_, ok := logger.(*ConsoleLogger)
			Expect(ok).To(BeTrue())
		})

		It("should return error when LOG_DESTINATION=api without GCP_PROJECT_ID", func() {
			os.Setenv(EnvLogDestination, "api")
			os.Unsetenv(EnvGCPProjectID)
			_, err := New()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("GCP_PROJECT_ID"))
		})

		It("should return error when LOG_DESTINATION=console-and-api without GCP_PROJECT_ID", func() {
			os.Setenv(EnvLogDestination, "console-and-api")
			os.Unsetenv(EnvGCPProjectID)
			_, err := New()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("GCP_PROJECT_ID"))
		})

		It("should return error for invalid LOG_LEVEL", func() {
			os.Setenv(EnvLogDestination, "console")
			os.Setenv(EnvLogLevel, "verbose")
			_, err := New()
			Expect(err).To(HaveOccurred())
		})
	})
})
