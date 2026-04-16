package logger

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"

	"cloud.google.com/go/logging"
)

var _ = Describe("GCP Logger", func() {

	// ---------------------------------------------------------------
	// mapSeverity — zap level to GCP severity mapping
	// ---------------------------------------------------------------
	Describe("mapSeverity", func() {
		DescribeTable("should map zap levels to GCP severities",
			func(zapLevel zapcore.Level, expected logging.Severity) {
				Expect(mapSeverity(zapLevel)).To(Equal(expected))
			},
			Entry("Debug → Debug", zapcore.DebugLevel, logging.Debug),
			Entry("Info → Info", zapcore.InfoLevel, logging.Info),
			Entry("Warn → Warning", zapcore.WarnLevel, logging.Warning),
			Entry("Error → Error", zapcore.ErrorLevel, logging.Error),
			Entry("DPanic → Critical", zapcore.DPanicLevel, logging.Critical),
			Entry("Panic → Critical", zapcore.PanicLevel, logging.Critical),
			Entry("Fatal → Critical", zapcore.FatalLevel, logging.Critical),
		)
	})

	// ---------------------------------------------------------------
	// gcpCore.Enabled — level filtering
	// ---------------------------------------------------------------
	Describe("gcpCore.Enabled", func() {
		It("should filter levels below the configured minimum", func() {
			core := &gcpCore{level: zapcore.InfoLevel}

			Expect(core.Enabled(zapcore.DebugLevel)).To(BeFalse())
			Expect(core.Enabled(zapcore.InfoLevel)).To(BeTrue())
			Expect(core.Enabled(zapcore.WarnLevel)).To(BeTrue())
			Expect(core.Enabled(zapcore.ErrorLevel)).To(BeTrue())
		})

		It("should pass all levels when set to Debug", func() {
			core := &gcpCore{level: zapcore.DebugLevel}

			Expect(core.Enabled(zapcore.DebugLevel)).To(BeTrue())
			Expect(core.Enabled(zapcore.InfoLevel)).To(BeTrue())
		})
	})

	// ---------------------------------------------------------------
	// gcpCore.With — field accumulation
	// ---------------------------------------------------------------
	Describe("gcpCore.With", func() {
		It("should return a new core with accumulated fields", func() {
			original := &gcpCore{level: zapcore.InfoLevel}

			child := original.With([]zapcore.Field{
				zapcore.Field{Key: "request_id", Type: zapcore.StringType, String: "abc-123"},
			})

			// Original should not be modified.
			gcpChild, ok := child.(*gcpCore)
			Expect(ok).To(BeTrue())
			Expect(gcpChild.fields).To(HaveLen(1))
			Expect(original.fields).To(BeEmpty())
		})
	})

	// ---------------------------------------------------------------
	// Factory — API destination env var validation
	// ---------------------------------------------------------------
	Describe("Factory with api destination", func() {
		AfterEach(func() {
			os.Unsetenv(EnvLogDestination)
			os.Unsetenv(EnvGCPProjectID)
			os.Unsetenv(EnvGCPLogName)
		})

		It("should return error when GCP_PROJECT_ID is missing", func() {
			os.Setenv(EnvLogDestination, "api")
			os.Unsetenv(EnvGCPProjectID)

			_, err := New()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("GCP_PROJECT_ID"))
		})

		It("should return error when project does not exist or is not accessible", func() {
			os.Setenv(EnvLogDestination, "api")
			os.Setenv(EnvGCPProjectID, "this-project-does-not-exist-abc123")

			_, err := New()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("GCP logging client has no access to project"))
		})
	})
})
