package vcsv

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func convertToType(value string, field reflect.StructField, tagOpts tagOptions) (reflect.Value, error) {
	t := field.Type
	if t == nil {
		return reflect.Value{}, fmt.Errorf("invalid field provided")
	}

	var err error
	var result interface{}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result, err = strconv.ParseInt(value, 10, t.Bits())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result, err = strconv.ParseUint(value, 10, t.Bits())
	case reflect.Float32, reflect.Float64:
		result, err = strconv.ParseFloat(value, t.Bits())
	case reflect.Complex64, reflect.Complex128:
		result, err = strconv.ParseComplex(value, t.Bits())
	case reflect.Bool:
		result, err = strconv.ParseBool(value)
	case reflect.String:
		result = value
	default:
		return convertByTypes(value, t, tagOpts)
	}

	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(result).Convert(t), nil
}

func convertByTypes(value string, fieldType reflect.Type, tagOpts tagOptions) (reflect.Value, error) {
	var err error
	var result interface{}

	switch fieldType {
	case reflect.TypeOf(time.Time{}):
		result, err = time.Parse(tagOpts.format, value)
	default:
		return convertTextUnmarshalerType(value, fieldType)
	}

	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(result).Convert(fieldType), nil
}

func convertTextUnmarshalerType(value string, fieldType reflect.Type) (reflect.Value, error) {
	targetType := determineTargetType(fieldType)

	unmarshalerType := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	if reflect.PtrTo(targetType).Implements(unmarshalerType) {
		return handleUnmarshalerConversion(value, fieldType, targetType)
	}

	return reflect.Value{}, fmt.Errorf("unsupported type %s", fieldType.Kind())
}

func determineTargetType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

func handleUnmarshalerConversion(value string, fieldType, targetType reflect.Type) (reflect.Value, error) {
	if value == "" {
		return handleEmptyValue(fieldType, targetType), nil
	}

	trimmedValue := strings.TrimSpace(value)
	ptr := reflect.New(targetType)

	if err := ptr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(trimmedValue)); err != nil {
		return reflect.Value{}, err
	}

	if fieldType.Kind() == reflect.Ptr {
		return ptr, nil
	}
	return ptr.Elem(), nil
}

func handleEmptyValue(fieldType, targetType reflect.Type) reflect.Value {
	if fieldType.Kind() == reflect.Struct {
		return reflect.Zero(fieldType)
	}
	return reflect.Zero(reflect.PtrTo(targetType))
}
