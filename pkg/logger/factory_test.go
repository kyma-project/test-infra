package logger

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
)

var _ = Describe("Factory", func() {

	Describe("New", func() {
		It("should create a ConsoleLogger when Destination is console", func() {
			cfg := Config{
				Level:       zapcore.InfoLevel,
				Destination: "console",
			}
			logger, err := New(context.Background(), cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger).NotTo(BeNil())

			_, ok := logger.(*ConsoleLogger)
			Expect(ok).To(BeTrue())
		})

		It("should default to console when Destination is empty", func() {
			cfg := Config{
				Level: zapcore.InfoLevel,
			}
			logger, err := New(context.Background(), cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger).NotTo(BeNil())

			_, ok := logger.(*ConsoleLogger)
			Expect(ok).To(BeTrue())
		})

		It("should return error when Destination is api without ProjectID", func() {
			cfg := Config{
				Level:       zapcore.InfoLevel,
				Destination: "api",
				ProjectID:   "",
			}
			_, err := New(context.Background(), cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ProjectID"))
		})

		It("should return error when Destination is console-and-api without ProjectID", func() {
			cfg := Config{
				Level:       zapcore.InfoLevel,
				Destination: "console-and-api",
				ProjectID:   "",
			}
			_, err := New(context.Background(), cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ProjectID"))
		})

		It("should return error for invalid Destination", func() {
			cfg := Config{
				Level:       zapcore.InfoLevel,
				Destination: "kafka",
			}
			_, err := New(context.Background(), cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("kafka"))
		})
	})

	Describe("Config env var constants", func() {
		It("should expose env var names for callers to use", func() {
			// Verify constants are exported and have expected values
			Expect(EnvLogDestination).To(Equal("LOG_DESTINATION"))
			Expect(EnvLogLevel).To(Equal("LOG_LEVEL"))
			Expect(EnvGCPProjectID).To(Equal("GCP_PROJECT_ID"))
			Expect(EnvGCPLogName).To(Equal("GCP_LOG_NAME"))
		})

		It("should allow caller to build Config from env vars", func() {
			os.Setenv(EnvLogDestination, "console")
			os.Setenv(EnvLogLevel, "debug")
			defer os.Unsetenv(EnvLogDestination)
			defer os.Unsetenv(EnvLogLevel)

			cfg := Config{
				Destination: os.Getenv(EnvLogDestination),
				Level:       zapcore.DebugLevel,
			}
			logger, err := New(context.Background(), cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger).NotTo(BeNil())
		})
	})
})
