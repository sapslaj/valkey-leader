package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

func Get[T any](name string) (T, error) {
	var value T
	var found bool

	raw, found := os.LookupEnv(name)
	if !found {
		return value, NewErrVarNotFound(name)
	}
	reflectValue := reflect.ValueOf(&value)
	elem := reflectValue.Elem()

	switch any(value).(type) {
	case string:
		elem.SetString(raw)
	case int:
		valueInt, err := strconv.ParseInt(raw, 10, 0)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetInt(valueInt)
	case int8:
		valueInt, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetInt(valueInt)
	case int16:
		valueInt, err := strconv.ParseInt(raw, 10, 16)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetInt(valueInt)
	case int32:
		valueInt, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetInt(valueInt)
	case int64:
		valueInt, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetInt(valueInt)
	case uint8:
		valueUint, err := strconv.ParseUint(raw, 10, 8)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetUint(valueUint)
	case uint16:
		valueUint, err := strconv.ParseUint(raw, 10, 16)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetUint(valueUint)
	case uint32:
		valueUint, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetUint(valueUint)
	case uint64:
		valueUint, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetUint(valueUint)
	case uint:
		valueUint, err := strconv.ParseUint(raw, 10, 0)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetUint(valueUint)
	case float32:
		valueFloat, err := strconv.ParseFloat(raw, 32)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetFloat(valueFloat)
	case float64:
		valueFloat, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetFloat(valueFloat)
	case bool:
		valueBool, err := strconv.ParseBool(raw)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetBool(valueBool)
	case time.Duration:
		valueDuration, err := time.ParseDuration(raw)
		if err != nil {
			return value, NewErrParsingWrapped(name, err)
		}
		elem.SetInt(int64(valueDuration))
	default:
		return value, NewErrUnsupportedType(name)
	}
	return value, nil
}

func MustGet[T any](name string) T {
	value, err := Get[T](name)
	if err != nil {
		panic(err)
	}
	return value
}

func GetDefault[T any](name string, defaultValue T) (T, error) {
	value, err := Get[T](name)
	if err != nil && IsErrVarNotFound(err) {
		return defaultValue, nil
	}
	return value, err
}

func MustGetDefault[T any](name string, defaultValue T) T {
	value, err := Get[T](name)
	if err != nil {
		return defaultValue
	}
	return value
}

type ErrVarNotFound struct {
	Name string
}

func (err *ErrVarNotFound) Error() string {
	return fmt.Sprintf("environment variable not found: %s", err.Name)
}

func NewErrVarNotFound(name string) *ErrVarNotFound {
	return &ErrVarNotFound{
		Name: name,
	}
}

func IsErrVarNotFound(err error) bool {
	_, ok := err.(*ErrVarNotFound)
	return ok
}

type ErrParsing struct {
	Name       string
	InnerError error
}

func (err *ErrParsing) Error() string {
	if err.InnerError == nil {
		return fmt.Sprintf("error parsing environment variable %s", err.Name)
	} else {
		return fmt.Sprintf("error parsing environment variable %s: %s", err.Name, err.InnerError.Error())
	}
}

func (err *ErrParsing) Unwrap() error {
	return err.InnerError
}

func (err *ErrParsing) Wrap(wrapped error) *ErrParsing {
	return &ErrParsing{
		Name:       err.Name,
		InnerError: wrapped,
	}
}

func NewErrParsing(name string) *ErrParsing {
	return &ErrParsing{
		Name: name,
	}
}

func NewErrParsingWrapped(name string, wrapped error) *ErrParsing {
	return &ErrParsing{
		Name:       name,
		InnerError: wrapped,
	}
}

func IsErrParsing(err error) bool {
	_, ok := err.(*ErrParsing)
	return ok
}

type ErrUnsupportedType struct {
	Name  string
	Value any
}

func (err *ErrUnsupportedType) Error() string {
	return fmt.Sprintf("unsupported type: %s", reflect.TypeOf(err.Value))
}

func NewErrUnsupportedType(name string) *ErrUnsupportedType {
	return &ErrUnsupportedType{
		Name: name,
	}
}

func IsErrUnsupportedType(err error) bool {
	_, ok := err.(*ErrUnsupportedType)
	return ok
}
