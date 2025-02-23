# otelzap [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Go Report Card][reportcard-img]][reportcard]

OpenTelemetry integration with ZAP logger.

Released under the [MIT License](LICENSE).


## Quick start

This package can be used to add regular log messages as OpenTelemetry events.

```{.go}
func MyAction(ctx context.Context, foo int) {
    ctx, span := tracer.Start(ctx, "MyAction")
    defer span.End()

    LOG := otelzap.SpanLogger(span, logger)

    LOG.Debug("do my action", zap.Int("foo", foo))
    // will write to stdout as normal zap logger does
    // and also will add "do my action" event to the span
}
```




[doc-img]: https://godoc.org/github.com/Pilatuz/otelzap?status.svg
[doc]: https://godoc.org/github.com/Pilatuz/otelzap
[ci-img]: https://github.com/Pilatuz/otelzap/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/Pilatuz/otelzap/actions
[reportcard-img]: https://goreportcard.com/badge/github.com/Pilatuz/otelzap
[reportcard]: https://goreportcard.com/report/github.com/Pilatuz/otelzap
