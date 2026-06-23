package logger

import (
	"context"
	"os"
	"time"

	gcplogging "cloud.google.com/go/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// captureSink is a test double for gcpLogSink that captures logged entries.
type captureSink struct {
	entries []gcplogging.Entry
}

func (s *captureSink) Log(e gcplogging.Entry) {
	s.entries = append(s.entries, e)
}

func (s *captureSink) Flush() error {
	return nil
}

func (s *captureSink) lastEntry() gcplogging.Entry {
	Expect(s.entries).NotTo(BeEmpty(), "expected at least one logged entry")
	return s.entries[len(s.entries)-1]
}

var _ = Describe("GCP Logger", func() {

	// ---------------------------------------------------------------
	// mapSeverity — zap level to GCP severity mapping
	// ---------------------------------------------------------------
	Describe("mapSeverity", func() {
		DescribeTable("should map zap levels to GCP severities",
			func(zapLevel zapcore.Level, expected gcplogging.Severity) {
				Expect(mapSeverity(zapLevel)).To(Equal(expected))
			},
			Entry("Debug → Debug", zapcore.DebugLevel, gcplogging.Debug),
			Entry("Info → Info", zapcore.InfoLevel, gcplogging.Info),
			Entry("Warn → Warning", zapcore.WarnLevel, gcplogging.Warning),
			Entry("Error → Error", zapcore.ErrorLevel, gcplogging.Error),
			Entry("DPanic → Critical", zapcore.DPanicLevel, gcplogging.Critical),
			Entry("Panic → Critical", zapcore.PanicLevel, gcplogging.Critical),
			Entry("Fatal → Critical", zapcore.FatalLevel, gcplogging.Critical),
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
			cfg := Config{
				Level:       "info",
				Destination: "api",
				ProjectID:   "",
			}
			_, err := New(context.Background(), cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ProjectID"))
		})
	})

	// ---------------------------------------------------------------
	// gcpCore.Write — trace field extraction
	// ---------------------------------------------------------------
	Describe("gcpCore.Write", func() {

		var (
			sink *captureSink
			core *gcpCore
		)

		BeforeEach(func() {
			sink = &captureSink{}
			core = &gcpCore{
				level: zapcore.DebugLevel,
				sink:  sink,
			}
		})

		writeEntry := func(c *gcpCore, msg string, fields ...zapcore.Field) {
			entry := zapcore.Entry{
				Level:   zapcore.InfoLevel,
				Time:    time.Now(),
				Message: msg,
			}
			err := c.Write(entry, fields)
			Expect(err).NotTo(HaveOccurred())
		}

		It("should extract logging.googleapis.com/trace into Entry.Trace", func() {
			coreWithTrace := core.With([]zapcore.Field{
				zap.String("logging.googleapis.com/trace", "projects/my-project/traces/abc123"),
			}).(*gcpCore)

			writeEntry(coreWithTrace, "test")
			e := sink.lastEntry()

			Expect(e.Trace).To(Equal("projects/my-project/traces/abc123"))
			// Should be removed from payload.
			payload := e.Payload.(map[string]interface{})
			Expect(payload).NotTo(HaveKey("logging.googleapis.com/trace"))
		})

		It("should extract logging.googleapis.com/spanId into Entry.SpanID", func() {
			coreWithSpan := core.With([]zapcore.Field{
				zap.String("logging.googleapis.com/spanId", "00f067aa0ba902b7"),
			}).(*gcpCore)

			writeEntry(coreWithSpan, "test")
			e := sink.lastEntry()

			Expect(e.SpanID).To(Equal("00f067aa0ba902b7"))
			payload := e.Payload.(map[string]interface{})
			Expect(payload).NotTo(HaveKey("logging.googleapis.com/spanId"))
		})

		It("should extract logging.googleapis.com/trace_sampled into Entry.TraceSampled", func() {
			coreWithSampled := core.With([]zapcore.Field{
				zap.Bool("logging.googleapis.com/trace_sampled", true),
			}).(*gcpCore)

			writeEntry(coreWithSampled, "test")
			e := sink.lastEntry()

			Expect(e.TraceSampled).To(BeTrue())
			payload := e.Payload.(map[string]interface{})
			Expect(payload).NotTo(HaveKey("logging.googleapis.com/trace_sampled"))
		})

		It("should set TraceSampled to false when field is false", func() {
			coreWithNotSampled := core.With([]zapcore.Field{
				zap.Bool("logging.googleapis.com/trace_sampled", false),
			}).(*gcpCore)

			writeEntry(coreWithNotSampled, "test")
			e := sink.lastEntry()

			Expect(e.TraceSampled).To(BeFalse())
		})

		It("should extract all trace fields together", func() {
			coreWithAll := core.With([]zapcore.Field{
				zap.String("logging.googleapis.com/trace", "projects/p/traces/t1"),
				zap.String("logging.googleapis.com/spanId", "span1"),
				zap.Bool("logging.googleapis.com/trace_sampled", true),
			}).(*gcpCore)

			writeEntry(coreWithAll, "full trace")
			e := sink.lastEntry()

			Expect(e.Trace).To(Equal("projects/p/traces/t1"))
			Expect(e.SpanID).To(Equal("span1"))
			Expect(e.TraceSampled).To(BeTrue())

			payload := e.Payload.(map[string]interface{})
			Expect(payload).NotTo(HaveKey("logging.googleapis.com/trace"))
			Expect(payload).NotTo(HaveKey("logging.googleapis.com/spanId"))
			Expect(payload).NotTo(HaveKey("logging.googleapis.com/trace_sampled"))
			Expect(payload).To(HaveKeyWithValue("message", "full trace"))
		})

		It("should keep regular fields in payload", func() {
			writeEntry(core, "hello", zap.String("user_id", "u1"), zap.Int("status", 200))
			e := sink.lastEntry()

			payload := e.Payload.(map[string]interface{})
			Expect(payload).To(HaveKeyWithValue("user_id", "u1"))
			Expect(payload).To(HaveKeyWithValue("message", "hello"))
		})

		It("should leave Entry.Trace empty when field is not present", func() {
			writeEntry(core, "no trace")
			e := sink.lastEntry()

			Expect(e.Trace).To(BeEmpty())
			Expect(e.SpanID).To(BeEmpty())
			Expect(e.TraceSampled).To(BeFalse())
		})
	})
})
