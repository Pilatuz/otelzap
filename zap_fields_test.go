package otelzap

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NotJSON returns an error on JSON marshaling.
type NotJSON struct {
	foo string
}

func (NotJSON) MarshalJSON() ([]byte, error) {
	return nil, assert.AnError
}

// Text is used to check TestMashaler interface.
type Text struct {
	foo string
}

func (text Text) MarshalText() ([]byte, error) {
	return []byte(text.foo), nil
}

// Stringer is used to check fmt.Stringer interface.
type Stringer struct {
	foo string
}

func (text Stringer) String() string {
	return text.foo
}

// TestAny unit tests for ZAP field conversion.
func TestAny(t *testing.T) {
	type (
		Bool    bool
		Int64   int64
		Uint64  uint64
		Float64 float64
		String  string

		Fixed    [2]int
		Int64s   []Int64
		Uint64s  []Uint64
		Bools    []Bool
		Float64s []Float64
		Strings  []String
	)

	assert.Equal(t, attribute.String("nil", "<nil>"), Any("nil", nil))
	assert.Equal(t, attribute.Bool("true", true), Any("true", true))
	assert.Equal(t, attribute.Bool("false", false), Any("false", false))
	assert.Equal(t, attribute.Bool("true", true), Any("true", Bool(true)))     // via reflection
	assert.Equal(t, attribute.Bool("false", false), Any("false", Bool(false))) // via reflection
	assert.Equal(t, attribute.BoolSlice("bools", []bool{true, false}), Any("bools", []bool{true, false}))
	assert.Equal(t, attribute.BoolSlice("bools", []bool{true, false}), Any("bools", Bools{true, false}))

	assert.Equal(t, attribute.String("foo", "foo"), Any("foo", "foo"))
	assert.Equal(t, attribute.StringSlice("strs", []string{"foo", "bar"}), Any("strs", []string{"foo", "bar"}))
	assert.Equal(t, attribute.StringSlice("strs", []string{"foo", "bar"}), Any("strs", Strings{"foo", "bar"}))
	assert.Equal(t, attribute.String("foo", "foo"), Any("foo", String("foo"))) // via reflection
	assert.Equal(t, attribute.String("base64", "AQIDBA=="), Any("base64", []byte{1, 2, 3, 4}))

	assert.Equal(t, attribute.Int("foo", 123), Any("foo", 123))
	assert.Equal(t, attribute.IntSlice("ints", []int{123, 456}), Any("ints", []int{123, 456}))
	assert.Equal(t, attribute.Int64Slice("ints", []int64{123, 456}), Any("ints", Fixed{123, 456}))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", int8(123)))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", int16(123)))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", int32(123)))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", int64(123)))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", Int64(123))) // via reflection
	assert.Equal(t, attribute.Int64Slice("ints", []int64{123, 456}), Any("ints", []int64{123, 456}))
	assert.Equal(t, attribute.Int64Slice("ints", []int64{123, 456}), Any("ints", Int64s{123, 456}))
	assert.Equal(t, attribute.Int64Slice("ints", []int64{123, 456}), Any("ints", Uint64s{123, 456}))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", uint(123)))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", uint8(123)))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", uint16(123)))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", uint32(123)))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", uint64(123)))
	assert.Equal(t, attribute.Int64("foo", 123), Any("foo", Uint64(123))) // via reflection

	assert.Equal(t, attribute.Float64("foo", 123), Any("foo", float32(123)))
	assert.Equal(t, attribute.Float64("foo", 123.456), Any("foo", 123.456))
	assert.Equal(t, attribute.Float64("foo", 123.456), Any("foo", Float64(123.456))) // via reflection
	assert.Equal(t, attribute.Float64Slice("floats", []float64{123.1, 456.2}), Any("floats", []float64{123.1, 456.2}))
	assert.Equal(t, attribute.Float64Slice("floats", []float64{123.1, 456.2}), Any("floats", Float64s{123.1, 456.2}))

	assert.Equal(t, attribute.String("stringer", "hello"), Any("stringer", Stringer{"hello"}))
	assert.Equal(t, attribute.String("array", `["foo","bar",123]`), Any("array", []interface{}{"foo", "bar", 123}))
	assert.Equal(t, attribute.String("object", `{"foo":"bar"}`), Any("object", map[string]interface{}{"foo": "bar"}))
	assert.Equal(t, attribute.String("not_json", `{hello}`), Any("not_json", NotJSON{"hello"})) // not json-convertible
	assert.Equal(t, attribute.String("text", `hello`), Any("text", Text{"hello"}))
}

// TestAppendZapField unit tests for appendZapField.
func TestAppendZapField(t *testing.T) {
	var (
		b = true
		// 	uptr uintptr    = 128
		c128 complex128 = 1.128 + 2.128i
		c64  complex64  = 1.64 + 2.64i
		bin  []byte
		t1   = time.Now()
		t2   time.Time
		foo  struct {
			Foo int
		}
		arr zapcore.ArrayMarshaler
		obj zapcore.ObjectMarshaler
	)

	assert.Nil(t, appendZapField(nil, zap.Skip()))
	assert.Nil(t, appendZapField(nil, zap.Namespace("ns")))
	assert.Equal(t, []attribute.KeyValue{attribute.Bool("bool", true)}, appendZapField(nil, zap.Bool("bool", true)))
	assert.Equal(t, []attribute.KeyValue{attribute.Bool("bool", true)}, appendZapField(nil, zap.Boolp("bool", &b)))
	assert.Equal(t, []attribute.KeyValue{attribute.String("bool", "<nil>")}, appendZapField(nil, zap.Boolp("bool", nil)))
	assert.Equal(t, []attribute.KeyValue{attribute.BoolSlice("bool", []bool{true, false})}, appendZapField(nil, zap.Bools("bool", []bool{true, false})))

	assert.Equal(t, []attribute.KeyValue{attribute.Float64("float32", 132.0)}, appendZapField(nil, zap.Float32("float32", 132.0)))
	assert.Equal(t, []attribute.KeyValue{attribute.Float64("float64", 1.64)}, appendZapField(nil, zap.Float64("float64", 1.64)))
	assert.Equal(t, []attribute.KeyValue{attribute.String("complex128", `(1.128E+00+2.128E+00i)`)}, appendZapField(nil, zap.Complex128("complex128", c128)))
	assert.Equal(t, []attribute.KeyValue{attribute.String("complex64", `(1.64E+00+2.64E+00i)`)}, appendZapField(nil, zap.Complex64("complex64", c64)))

	assert.Equal(t, []attribute.KeyValue{attribute.Int64("int", 100)}, appendZapField(nil, zap.Int("int", 100)))
	assert.Equal(t, []attribute.KeyValue{attribute.Int64("int64", 64)}, appendZapField(nil, zap.Int64("int64", 64)))
	assert.Equal(t, []attribute.KeyValue{attribute.Int64("int32", 32)}, appendZapField(nil, zap.Int32("int32", 32)))
	assert.Equal(t, []attribute.KeyValue{attribute.Int64("int16", 16)}, appendZapField(nil, zap.Int32("int16", 16)))
	assert.Equal(t, []attribute.KeyValue{attribute.Int64("int8", 8)}, appendZapField(nil, zap.Int32("int8", 8)))
	assert.Equal(t, []attribute.KeyValue{attribute.Int64("uint", 100)}, appendZapField(nil, zap.Uint("uint", 100)))
	assert.Equal(t, []attribute.KeyValue{attribute.Int64("uint64", 64)}, appendZapField(nil, zap.Uint64("uint64", 64)))
	assert.Equal(t, []attribute.KeyValue{attribute.Int64("uint32", 32)}, appendZapField(nil, zap.Uint32("uint32", 32)))
	assert.Equal(t, []attribute.KeyValue{attribute.Int64("uint16", 16)}, appendZapField(nil, zap.Uint32("uint16", 16)))
	assert.Equal(t, []attribute.KeyValue{attribute.Int64("uint8", 8)}, appendZapField(nil, zap.Uint32("uint8", 8)))

	assert.Equal(t, []attribute.KeyValue{attribute.String("string", "hello")}, appendZapField(nil, zap.String("string", "hello")))
	assert.Equal(t, []attribute.KeyValue{attribute.String("stringer", "1ms")}, appendZapField(nil, zap.Stringer("stringer", time.Millisecond)))
	assert.Equal(t, []attribute.KeyValue{attribute.String("binary", "")}, appendZapField(nil, zap.Binary("binary", bin)))
	assert.Equal(t, []attribute.KeyValue{attribute.String("byte_string", "hello")}, appendZapField(nil, zap.ByteString("byte_string", []byte("hello"))))
	assert.Equal(t, []attribute.KeyValue{attribute.String("error", "assert.AnError general error for testing")}, appendZapField(nil, zap.Error(assert.AnError)))

	assert.Equal(t, []attribute.KeyValue{attribute.String("duration", "1ms")}, appendZapField(nil, zap.Duration("duration", time.Millisecond)))
	assert.Equal(t, []attribute.KeyValue{attribute.String("time1", t1.Format(time.RFC3339Nano))}, appendZapField(nil, zap.Time("time1", t1)))
	assert.Equal(t, []attribute.KeyValue{attribute.String("time2", t2.Format(time.RFC3339Nano))}, appendZapField(nil, zap.Time("time2", t2)))

	assert.Equal(t, []attribute.KeyValue{attribute.String("reflect", `{"Foo":0}`)}, appendZapField(nil, zap.Reflect("reflect", foo)))
	assert.Equal(t, []attribute.KeyValue{attribute.String("array", `<nil>`)}, appendZapField(nil, zap.Array("array", arr)))
	assert.Equal(t, []attribute.KeyValue{attribute.String("object", `<nil>`)}, appendZapField(nil, zap.Object("object", obj)))
	assert.Equal(t, []attribute.KeyValue{attribute.String("", `<nil>`)}, appendZapField(nil, zap.Inline(obj)))
}

// TestAttributes unit tests for attributes.
func TestAttributes(t *testing.T) {
	assert.Nil(t, attributesFromZapFields(nil, nil))
	assert.Equal(t,
		[]attribute.KeyValue{
			attribute.String("foo", "hello"),
			attribute.Int("bar", 123),
		},
		attributesFromZapFields(nil, nil,
			attribute.String("foo", "hello"),
			attribute.Int("bar", 123),
		))

	assert.Equal(t,
		[]attribute.KeyValue{
			attribute.Int("foo", 111),
			attribute.Int("bar", 222),
			attribute.Int("baz", 333),
		},
		attributesFromZapFields(
			[]zapcore.Field{zap.Int("bar", 222)},
			[]zapcore.Field{zap.Int("baz", 333)},
			attribute.Int("foo", 111),
		))
	assert.Equal(t,
		[]attribute.KeyValue{
			attribute.Int("foo", 111),
			attribute.Int("baz", 333),
		},
		attributesFromZapFields(
			nil,
			[]zapcore.Field{zap.Int("baz", 333)},
			attribute.Int("foo", 111),
		))
	assert.Equal(t,
		[]attribute.KeyValue{
			attribute.Int("foo", 111),
			attribute.Int("bar", 222),
		},
		attributesFromZapFields(
			[]zapcore.Field{zap.Int("bar", 222)},
			nil,
			attribute.Int("foo", 111),
		))

	assert.Nil(t, AppendZapFields(nil))
	assert.Equal(t,
		[]attribute.KeyValue{
			attribute.Int("foo", 111),
			attribute.String("bar", "test"),
		},
		AppendZapFields(nil,
			zap.Int("foo", 111),
			zap.String("bar", "test")))
}

// TestHTTPHeader unit tests for HTTPHeader function.
func TestHTTPHeader(t *testing.T) {
	assert.Equal(t,
		attribute.String("foo", ""),
		HTTPHeader("foo", nil, nil))
	assert.Equal(t,
		attribute.String("foo", ""),
		HTTPHeader("foo", http.Header{}, nil))

	h := http.Header{}
	h.Add("a", "1")
	h.Add("b", "2")
	h.Add("a", "3")
	assert.Equal(t,
		attribute.String("foo", "A: 1\r\nA: 3\r\nB: 2\r\n"),
		HTTPHeader("foo", h, nil))

	h = http.Header{}
	h.Add("content-type", "application/json")
	h.Add("authorization", "Bearer ups")
	h.Add("content-length", "1024")
	assert.Equal(t,
		attribute.String("foo", "Content-Length: 1024\r\nContent-Type: application/json\r\n"),
		HTTPHeader("foo", h, map[string]bool{"Authorization": true}))
}

// TestConcat unit tests for concatFields function.
func TestConcat(t *testing.T) {
	assert.Nil(t, concatFields(nil, nil))
	assert.Nil(t, concatFields([]zapcore.Field{}, nil))
	assert.Empty(t, concatFields(nil, []zapcore.Field{}))
	assert.Empty(t, concatFields([]zapcore.Field{}, []zapcore.Field{}))

	assert.Equal(t, []zapcore.Field{zap.Int("a", 1)}, concatFields([]zapcore.Field{}, []zapcore.Field{zap.Int("a", 1)}))
	assert.Equal(t, []zapcore.Field{zap.Int("a", 1)}, concatFields([]zapcore.Field{zap.Int("a", 1)}, []zapcore.Field{}))
	assert.Equal(t, []zapcore.Field{zap.Int("a", 1), zap.Int("b", 2)}, concatFields([]zapcore.Field{zap.Int("a", 1)}, []zapcore.Field{zap.Int("b", 2)}))
}
