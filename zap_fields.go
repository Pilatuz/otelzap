package otelzap

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"
)

// attributesFromZapFields converts multiple ZAP fields into OpenTelemetry attributes.
func attributesFromZapFields(
	with []zapcore.Field,
	fields []zapcore.Field,
	extra ...attribute.KeyValue,
) []attribute.KeyValue {
	if len(with)+len(fields) == 0 {
		// no fields, use extra attributes only
		return extra
	}

	// convert each ZAP field...
	attrs := make([]attribute.KeyValue, 0, len(with)+len(fields)+len(extra))
	attrs = append(attrs, extra...) // use extra "as is"

	// convert "with" fields first
	for _, field := range with {
		attrs = appendZapField(attrs, field)
	}

	// then rest of fields
	for _, field := range fields {
		attrs = appendZapField(attrs, field)
	}

	return attrs
}

// appendZapField converts and appends a ZAP field.
func appendZapField(attributes []attribute.KeyValue, field zapcore.Field) []attribute.KeyValue {
	switch field.Type {
	case zapcore.SkipType, // see zap.Skip()
		zapcore.NamespaceType: // see zap.Namespace()
		return attributes // skip it

	case zapcore.BoolType: // see zap.Bool()
		return append(attributes, attribute.Bool(field.Key, field.Integer != 0))

	case zapcore.Int8Type, // see zap.Int8()
		zapcore.Int16Type,   // see zap.Int16()
		zapcore.Int32Type,   // see zap.Int32()
		zapcore.Int64Type,   // see zap.Int64()
		zapcore.Uint8Type,   // see zap.Uint8()
		zapcore.Uint16Type,  // see zap.Uint16()
		zapcore.Uint32Type,  // see zap.Uint32()
		zapcore.Uint64Type,  // see zap.Uint64()
		zapcore.UintptrType: // see zap.Uintptr()
		return append(attributes, attribute.Int64(field.Key, field.Integer))

	case zapcore.Float32Type: // see zap.Float32()
		return append(attributes, attribute.Float64(field.Key, float64(math.Float32frombits(uint32(field.Integer)))))
	case zapcore.Float64Type: // see zap.Float64()
		return append(attributes, attribute.Float64(field.Key, math.Float64frombits(uint64(field.Integer))))

	case zapcore.Complex64Type: // see zap.Complex64()
		s := strconv.FormatComplex(complex128(field.Interface.(complex64)), 'E', -1, 64)
		return append(attributes, attribute.String(field.Key, s))
	case zapcore.Complex128Type: // see zap.Complex128()
		s := strconv.FormatComplex(field.Interface.(complex128), 'E', -1, 128)
		return append(attributes, attribute.String(field.Key, s))

	case zapcore.StringType: // see zap.String()
		return append(attributes, attribute.String(field.Key, field.String))
	case zapcore.BinaryType: // see zap.Binary()
		return append(attributes, attribute.String(field.Key, base64.StdEncoding.EncodeToString(field.Interface.([]byte))))
	case zapcore.ByteStringType: // see zap.ByteString()
		return append(attributes, attribute.String(field.Key, string(field.Interface.([]byte))))
	case zapcore.StringerType: // see zap.Stringer()
		return append(attributes, attribute.Stringer(field.Key, field.Interface.(fmt.Stringer)))

	case zapcore.DurationType: // see zap.Duration()
		return append(attributes, attribute.Stringer(field.Key, time.Duration(field.Integer)))
	case zapcore.TimeType: // see zap.Time()
		t := time.Unix(0, field.Integer).In(field.Interface.(*time.Location))
		return append(attributes, attribute.String(field.Key, t.Format(time.RFC3339Nano)))
	case zapcore.TimeFullType: // see zap.Time()
		return append(attributes, attribute.String(field.Key, field.Interface.(time.Time).Format(time.RFC3339Nano)))

	case zapcore.ErrorType: // see zap.Error()
		return append(attributes, attribute.String(field.Key, field.Interface.(error).Error()))

	case zapcore.ReflectType, // see zap.Reflect()
		zapcore.ArrayMarshalerType,  // see zap.Strings(), zap.Int64s(), ...
		zapcore.ObjectMarshalerType, // see zap.Object()
		zapcore.InlineMarshalerType: // see zap.Inline()
		break // return append(attributes, Any(field.Key, field.Interface))
	}

	return append(attributes, Any(field.Key, field.Interface))
}

// HTTPHeader converts HTTP headers into OpenTelemetry attribute as multi-line string.
// The HTTP headers to exclude should be in cacnonical form (see textproto.CanonicalMIMEHeaderKey).
func HTTPHeader(key string, header http.Header, exclude ...string) attribute.KeyValue {
	var buf bytes.Buffer
	if err := header.WriteSubset(&buf, excludeMap(exclude)); err != nil { // unlikely
		return attribute.String(key, err.Error())
	}
	return attribute.String(key, buf.String())
}

// excludeMap build exclusion map
func excludeMap(keys []string) map[string]bool {
	if len(keys) == 0 {
		return nil // empty
	}

	out := make(map[string]bool, len(keys))
	for _, k := range keys {
		out[k] = true
	}
	return out
}

// Any converts unknown type to OpenTelemetry attribute, probably as JSON value.
func Any(key string, value interface{}) attribute.KeyValue {
	switch t := value.(type) {
	case nil:
		return attribute.String(key, "<nil>")

	case bool:
		return attribute.Bool(key, t)
	case []bool:
		return attribute.BoolSlice(key, t)

	case string:
		return attribute.String(key, t)
	case []string:
		return attribute.StringSlice(key, t)
	case []byte:
		return attribute.String(key, base64.StdEncoding.EncodeToString(t))

	case int:
		return attribute.Int(key, t)
	case []int:
		return attribute.IntSlice(key, t)

	case int8:
		return attribute.Int64(key, int64(t))
	case int16:
		return attribute.Int64(key, int64(t))
	case int32:
		return attribute.Int64(key, int64(t))
	case int64:
		return attribute.Int64(key, t)
	case []int64:
		return attribute.Int64Slice(key, t)

	case uint:
		return attribute.Int64(key, int64(t))
	case uint8:
		return attribute.Int64(key, int64(t))
	case uint16:
		return attribute.Int64(key, int64(t))
	case uint32:
		return attribute.Int64(key, int64(t))
	case uint64:
		return attribute.Int64(key, int64(t))

	case float32:
		return attribute.Float64(key, float64(t))
	case float64:
		return attribute.Float64(key, t)
	case []float64:
		return attribute.Float64Slice(key, t)

	case encoding.TextMarshaler:
		if b, err := t.MarshalText(); err == nil {
			return attribute.String(key, string(b))
		}
		// in case of error just try something else below
	case fmt.Stringer:
		return attribute.Stringer(key, t)
	}

	// try reflected value
	switch rv := reflect.ValueOf(value); rv.Kind() {
	case reflect.Bool:
		return attribute.Bool(key, rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return attribute.Int64(key, rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return attribute.Int64(key, int64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		return attribute.Float64(key, rv.Float())
	case reflect.String:
		return attribute.String(key, rv.String())

	case reflect.Slice, reflect.Array:
		switch rv.Type().Elem().Kind() {
		case reflect.Bool:
			return attribute.BoolSlice(key, toBoolSlice(rv))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return attribute.Int64Slice(key, toInt64Slice(rv))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return attribute.Int64Slice(key, toUint64Slice(rv))
		case reflect.Float64:
			return attribute.Float64Slice(key, toFloat64Slice(rv))
		case reflect.String:
			return attribute.StringSlice(key, toStringSlice(rv))
		}
	}

	// format as JSON
	if b, err := json.Marshal(value); err == nil {
		return attribute.String(key, string(b))
	}

	// format as %v string as a final option
	return attribute.String(key, fmt.Sprint(value))
}

// toBoolSlice converts reflected value to bool slice.
func toBoolSlice(rv reflect.Value) []bool {
	N := rv.Len()
	out := make([]bool, N)
	for i := 0; i < N; i++ {
		re := rv.Index(i)
		out[i] = re.Bool()
	}
	return out
}

// toInt64Slice converts reflected value to int64 slice.
func toInt64Slice(rv reflect.Value) []int64 {
	N := rv.Len()
	out := make([]int64, N)
	for i := 0; i < N; i++ {
		re := rv.Index(i)
		out[i] = re.Int()
	}
	return out
}

// toUint64Slice converts reflected value to int64 slice.
func toUint64Slice(rv reflect.Value) []int64 {
	N := rv.Len()
	out := make([]int64, N)
	for i := 0; i < N; i++ {
		re := rv.Index(i)
		out[i] = int64(re.Uint())
	}
	return out
}

// toFloat64Slice converts reflected value to float64 slice.
func toFloat64Slice(rv reflect.Value) []float64 {
	N := rv.Len()
	out := make([]float64, N)
	for i := 0; i < N; i++ {
		re := rv.Index(i)
		out[i] = re.Float()
	}
	return out
}

// toStringSlice converts reflected value to string slice.
func toStringSlice(rv reflect.Value) []string {
	N := rv.Len()
	out := make([]string, N)
	for i := 0; i < N; i++ {
		re := rv.Index(i)
		out[i] = re.String()
	}
	return out
}

// concatFields concatenates two set of fields.
func concatFields(a []zapcore.Field, b []zapcore.Field) []zapcore.Field {
	if len(a) == 0 {
		return b // first set is empty, return second one
	}
	if len(b) == 0 {
		return a // second set is empty, return first one
	}

	// concatenate both non-empty sets
	out := make([]zapcore.Field, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)

	return out
}
