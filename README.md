# otelzap [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Go Report Card][reportcard-img]][reportcard]

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
}
```

The code above will also add OpenTelemetry event with "foo" attribute attached.


[doc-img]: https://godoc.org/github.com/Pilatuz/otelzap?status.svg
[doc]: https://godoc.org/github.com/Pilatuz/otelzap
[ci-img]: https://github.com/Pilatuz/otelzap/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/Pilatuz/otelzap/actions
[cov-img]: https://codecov.io/gh/Pilatuz/otelzap/branch/main/graph/badge.svg
[cov]: https://codecov.io/gh/Pilatuz/otelzap
[reportcard-img]: https://goreportcard.com/badge/github.com/Pilatuz/otelzap
[reportcard]: https://goreportcard.com/report/github.com/Pilatuz/otelzap
