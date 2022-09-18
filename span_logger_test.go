package otelzap_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	. "github.com/Pilatuz/otelzap"
	. "github.com/Pilatuz/otelzap/internal/mocked"
)

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -mock_names Span=MockedSpan -package mocked -destination internal/mocked/mocked_span.go go.opentelemetry.io/otel/trace Span

func TestSpanLogger(t *testing.T) {
	L1 := zap.NewNop()
	assert.Same(t, L1, SpanLoggerFromContext(context.Background(), L1))

	ctrl := gomock.NewController(t)
	span := NewMockedSpan(ctrl)
	span.EXPECT().
		IsRecording().
		Return(true).
		AnyTimes()

	L2, buf2 := newJSONLogger()
	L2 = L2.Named("my")
	SL2 := SpanLogger(span, L2)
	SL2 = SL2.
		With(zap.String("bar", "hello")).
		With(zap.Int("baz", 321))

	span.EXPECT().
		AddEvent("my message",
			trace.WithAttributes(
				attribute.String("zap.level", "info"),
				attribute.String("zap.logger_name", "my"),
				attribute.String("bar", "hello"),
				attribute.Int("baz", 321),
				attribute.Int("foo", 123),
			))
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
