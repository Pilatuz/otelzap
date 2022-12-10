package otelzap

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SpanLogger creates ZAP logger which also writes to OpenTelemetry span.
// If span is `nil“ or `no-op` then the same logger returned.
func SpanLogger(span trace.Span, logger *zap.Logger) *zap.Logger {
	if span == nil || !span.IsRecording() {
		return logger // no tracing enabled
	}

	wrap := func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core,
			zapSpanCore{
				core: core,
				span: span,
			})
	}

	return logger.WithOptions(zap.WrapCore(wrap))
}

// SpanLoggerFromContext similar to SpanLogger but gets span from context.
func SpanLoggerFromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	return SpanLogger(trace.SpanFromContext(ctx), logger)
}

// zapSpanCore writes log entries to the span as OpenTelemetry events.
type zapSpanCore struct {
	core zapcore.Core // actually is used to check levels
	span trace.Span
	with []zapcore.Field
}

// Enabled checks if logging level is enabled.
func (zs zapSpanCore) Enabled(level zapcore.Level) bool {
	return zs.core.Enabled(level)
}

// With adds structured context to the Core.
func (zs zapSpanCore) With(fields []zapcore.Field) zapcore.Core {
	return zapSpanCore{
		core: zs.core, // zs.core.With(fields), - no sense yet
		span: zs.span,
		with: concatFields(zs.with, fields),
	}
}

// Check determines whether the supplied Entry should be logged.
func (zs zapSpanCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if zs.Enabled(entry.Level) {
		checked = checked.AddCore(entry, zs)
	}

	return checked
}

// Write serializes the Entry and any Fields supplied at the log site and
// writes them to OpenTelemetry as an event.
func (zs zapSpanCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	zs.span.AddEvent(entry.Message,
		trace.WithAttributes(attributesFromZapFields(zs.with, fields,
			attribute.Stringer("zap.level", entry.Level),
			attribute.String("zap.logger_name", entry.LoggerName),
		)...))

	return nil
}

// Sync flushes buffered logs.
func (zs zapSpanCore) Sync() error {
	return nil // nothing to sync
}
