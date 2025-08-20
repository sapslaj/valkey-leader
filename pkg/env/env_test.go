package env

import (
	"errors"
	"testing"
	"time"
)

// NOTE: yes I know Testify would be way better here but trying not to have any
// external deps
// If you would like to use Testify, import
// "github.com/stretchr/testify/assert" then run `sed 's/TAssert/assert./g`
func TAssertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("received unexpected error: %+v", err)
	}
}

func TAssertEqual(t *testing.T, expected any, actual any) {
	if expected != actual {
		t.Errorf("not equal: expected: %v ; got %v", expected, actual)
	}
}

func TAssertPanics(t *testing.T, f func()) {
	didPanic := func(f func()) (did bool) {
		did = true
		defer func() { recover() }()
		f()
		did = false
		return
	}

	if !didPanic(f) {
		t.Errorf("func did not panic when it was expected to")
	}
}

func TestGet(t *testing.T) {
	// string
	t.Setenv("TEST_STRING", "foo")
	vString, err := Get[string]("TEST_STRING")
	TAssertNoError(t, err)
	TAssertEqual(t, "foo", vString)

	// int
	t.Setenv("TEST_INT", "123456")
	vInt, err := Get[int]("TEST_INT")
	TAssertNoError(t, err)
	TAssertEqual(t, 123456, vInt)

	// int8
	t.Setenv("TEST_INT8", "1")
	vInt8, err := Get[int8]("TEST_INT8")
	TAssertNoError(t, err)
	TAssertEqual(t, int8(1), vInt8)

	// int16
	t.Setenv("TEST_INT16", "12345")
	vInt16, err := Get[int16]("TEST_INT16")
	TAssertNoError(t, err)
	TAssertEqual(t, int16(12345), vInt16)

	// int32
	t.Setenv("TEST_INT32", "1234567890")
	vInt32, err := Get[int32]("TEST_INT32")
	TAssertNoError(t, err)
	TAssertEqual(t, int32(1234567890), vInt32)

	// int64
	t.Setenv("TEST_INT64", "9223372036854775807")
	vInt64, err := Get[int64]("TEST_INT64")
	TAssertNoError(t, err)
	TAssertEqual(t, int64(9223372036854775807), vInt64)

	// uint8
	t.Setenv("TEST_UINT8", "255")
	vUint8, err := Get[uint8]("TEST_UINT8")
	TAssertNoError(t, err)
	TAssertEqual(t, uint8(255), vUint8)

	// uint16
	t.Setenv("TEST_UINT16", "65535")
	vUint16, err := Get[uint16]("TEST_UINT16")
	TAssertNoError(t, err)
	TAssertEqual(t, uint16(65535), vUint16)

	// uint32
	t.Setenv("TEST_UINT32", "4294967295")
	vUint32, err := Get[uint32]("TEST_UINT32")
	TAssertNoError(t, err)
	TAssertEqual(t, uint32(4294967295), vUint32)

	// uint64
	t.Setenv("TEST_UINT64", "18446744073709551615")
	vUint64, err := Get[uint64]("TEST_UINT64")
	TAssertNoError(t, err)
	TAssertEqual(t, uint64(18446744073709551615), vUint64)

	// uint
	t.Setenv("TEST_UINT", "4000000000")
	vUint, err := Get[uint]("TEST_UINT")
	TAssertNoError(t, err)
	TAssertEqual(t, uint(4000000000), vUint)

	// float32
	t.Setenv("TEST_FLOAT32", "3.14159")
	vFloat32, err := Get[float32]("TEST_FLOAT32")
	TAssertNoError(t, err)
	TAssertEqual(t, float32(3.14159), vFloat32)

	// float64
	t.Setenv("TEST_FLOAT64", "3.141592653589793")
	vFloat64, err := Get[float64]("TEST_FLOAT64")
	TAssertNoError(t, err)
	TAssertEqual(t, float64(3.141592653589793), vFloat64)

	// bool
	t.Setenv("TEST_BOOL_TRUE", "true")
	vBoolTrue, err := Get[bool]("TEST_BOOL_TRUE")
	TAssertNoError(t, err)
	TAssertEqual(t, true, vBoolTrue)

	t.Setenv("TEST_BOOL_FALSE", "false")
	vBoolFalse, err := Get[bool]("TEST_BOOL_FALSE")
	TAssertNoError(t, err)
	TAssertEqual(t, false, vBoolFalse)

	t.Setenv("TEST_DURATION", "5s")
	vDuration, err := Get[time.Duration]("TEST_DURATION")
	TAssertNoError(t, err)
	TAssertEqual(t, 5*time.Second, vDuration)

	// Error cases

	// Variable not found
	_, err = Get[string]("NONEXISTENT_VAR")
	if err == nil {
		t.Errorf("expected error but got nil")
	} else if !IsErrVarNotFound(err) {
		t.Errorf("error is not ErrVarNotFound")
	}

	// Parsing error
	t.Setenv("TEST_PARSE_ERROR_INT", "not_an_int")
	_, err = Get[int]("TEST_PARSE_ERROR_INT")
	if err == nil {
		t.Errorf("expected error but got nil")
	} else if !IsErrParsing(err) {
		t.Errorf("error is not ErrParsing")
	}

	// Unsupported type
	type CustomType struct{}
	_, err = Get[CustomType]("TEST_STRING")
	if err == nil {
		t.Errorf("expected error but got nil")
	} else if !IsErrUnsupportedType(err) {
		t.Errorf("error is not ErrUnsupportedType")
	}
}

func TestMustGet(t *testing.T) {
	// Successful retrieval
	t.Setenv("TEST_MUST_GET_STRING", "success")
	value := MustGet[string]("TEST_MUST_GET_STRING")
	TAssertEqual(t, "success", value)

	// Should panic when variable not found
	TAssertPanics(t, func() {
		MustGet[string]("NONEXISTENT_VAR_MUST_GET")
	})

	// Should panic when parsing fails
	t.Setenv("TEST_MUST_GET_PARSE_ERROR", "not_an_int")
	TAssertPanics(t, func() {
		MustGet[int]("TEST_MUST_GET_PARSE_ERROR")
	})
}

func TestGetDefault(t *testing.T) {
	// Variable exists - should return the value
	t.Setenv("TEST_GET_DEFAULT_EXISTS", "existing_value")
	value, err := GetDefault[string]("TEST_GET_DEFAULT_EXISTS", "default_value")
	TAssertNoError(t, err)
	TAssertEqual(t, "existing_value", value)

	// Variable doesn't exist - should return the default value
	value, err = GetDefault[string]("NONEXISTENT_VAR_GET_DEFAULT", "default_value")
	TAssertNoError(t, err)
	TAssertEqual(t, "default_value", value)

	// Parsing error - should return an error, not the default value
	t.Setenv("TEST_GET_DEFAULT_PARSE_ERROR", "not_an_int")
	_, err = GetDefault[int]("TEST_GET_DEFAULT_PARSE_ERROR", 42)
	if err == nil {
		t.Errorf("expected error but got nil")
	} else if !IsErrParsing(err) {
		t.Errorf("error is not ErrParsing")
	}

	// Test with various types
	t.Setenv("TEST_GET_DEFAULT_INT", "123")
	intVal, err := GetDefault[int]("TEST_GET_DEFAULT_INT", 0)
	TAssertNoError(t, err)
	TAssertEqual(t, 123, intVal)

	t.Setenv("TEST_GET_DEFAULT_BOOL", "true")
	boolVal, err := GetDefault[bool]("TEST_GET_DEFAULT_BOOL", false)
	TAssertNoError(t, err)
	TAssertEqual(t, true, boolVal)
}

func TestMustGetDefault(t *testing.T) {
	// Variable exists - should return the value
	t.Setenv("TEST_MUST_GET_DEFAULT_EXISTS", "existing_value")
	value := MustGetDefault[string]("TEST_MUST_GET_DEFAULT_EXISTS", "default_value")
	TAssertEqual(t, "existing_value", value)

	// Variable doesn't exist - should return the default value
	value = MustGetDefault[string]("NONEXISTENT_VAR_MUST_GET_DEFAULT", "default_value")
	TAssertEqual(t, "default_value", value)

	// Parsing error - should return the default value
	t.Setenv("TEST_MUST_GET_DEFAULT_PARSE_ERROR", "not_an_int")
	intValue := MustGetDefault[int]("TEST_MUST_GET_DEFAULT_PARSE_ERROR", 42)
	TAssertEqual(t, 42, intValue)

	// Test with various types
	t.Setenv("TEST_MUST_GET_DEFAULT_FLOAT", "3.14")
	floatVal := MustGetDefault[float64]("TEST_MUST_GET_DEFAULT_FLOAT", 0.0)
	TAssertEqual(t, 3.14, floatVal)

	t.Setenv("TEST_MUST_GET_DEFAULT_UINT", "255")
	uintVal := MustGetDefault[uint8]("TEST_MUST_GET_DEFAULT_UINT", uint8(0))
	TAssertEqual(t, uint8(255), uintVal)
}

func TestIsErrVarNotFound(t *testing.T) {
	t.Parallel()

	if IsErrVarNotFound(errors.New("other error")) {
		t.Error("should be false")
	}

	if !IsErrVarNotFound(NewErrVarNotFound("TEST")) {
		t.Error("should be true")
	}
}

func TestIsErrParsing(t *testing.T) {
	t.Parallel()

	if IsErrParsing(errors.New("other error")) {
		t.Error("should be false")
	}

	if !IsErrParsing(NewErrParsing("TEST")) {
		t.Error("should be true")
	}
}

func TestIsErrUnsupportedType(t *testing.T) {
	t.Parallel()

	if IsErrUnsupportedType(errors.New("other error")) {
		t.Error("should be false")
	}

	if !IsErrUnsupportedType(NewErrUnsupportedType("TEST")) {
		t.Error("should be true")
	}
}
