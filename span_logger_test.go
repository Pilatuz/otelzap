package otelzap_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	. "github.com/Pilatuz/otelzap"
)

// mockedSpan span for tests.
type mockedSpan struct {
	trace.Span

	addEventCb func(string, ...trace.EventOption)
}

func (mockedSpan) IsRecording() bool { return true }
func (m mockedSpan) AddEvent(name string, options ...trace.EventOption) {
	m.addEventCb(name, options...)
}

// TestSpanLogger unit tests for SpanLogger.
func TestSpanLogger(t *testing.T) {
	L1 := zap.NewNop()
	assert.Same(t, L1, SpanLoggerFromContext(context.Background(), L1))

	span := mockedSpan{
		Span: noop.Span{},
	}

	L2, buf2 := newJSONLogger()
	L2 = L2.Named("my")
	SL2 := SpanLogger(span, L2)
	SL2 = SL2.
		With(zap.String("bar", "hello")).
		With(zap.Int("baz", 321))

	span.addEventCb = func(name string, options ...trace.EventOption) {
		cfg := trace.NewEventConfig(options...)
		assert.Equal(t, "my message", name)
		assert.Equal(t,
			[]attribute.KeyValue{
				attribute.String("zap.level", "info"),
				attribute.String("zap.logger_name", "my"),
				attribute.String("bar", "hello"),
				attribute.Int("baz", 321),
				attribute.Int("foo", 123),
			}, cfg.Attributes())
	}
	SL2.Info("my message", zap.Int("foo", 123))
	SL2.Debug("my message", zap.String("foo", "ignore me"))

	assert.NoError(t, SL2.Sync())
	assert.Equal(t, `{"level":"info","msg":"my message","bar":"hello","baz":321,"foo":123}`, buf2.Stripped())
}

// newJSONLogger creates a new zap.Logger instance with a zaptest.Buffer as a writer.
func newJSONLogger() (*zap.Logger, *zaptest.Buffer) {
	encoder := zapcore.NewJSONEncoder(
		zapcore.EncoderConfig{
			MessageKey:     "msg",
			LevelKey:       "level",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		})
	buf := &zaptest.Buffer{}
	logger := zap.New(zapcore.NewCore(encoder, buf, zapcore.InfoLevel))
	return logger, buf
}
