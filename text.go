package reflectx

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"

	pkgerrors "github.com/pkg/errors"
)

func MarshalText(v interface{}) ([]byte, error) {
	if rv, ok := v.(reflect.Value); ok {
		for rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil, nil
			}
			rv = rv.Elem()
		}

		if rv.CanInterface() {
			v = rv.Interface()
		}
	}

	if textMarshaler, ok := v.(encoding.TextMarshaler); ok {
		return textMarshaler.MarshalText()
	}

	if v == nil {
		return nil, nil
	}

	switch x := v.(type) {
	case []byte:
		return x, nil
	case string:
		return []byte(x), nil
	case bool:
		return strconv.AppendBool([]byte{}, x), nil
	case int:
		return strconv.AppendInt([]byte{}, int64(x), 10), nil
	case int8:
		return strconv.AppendInt([]byte{}, int64(x), 10), nil
	case int16:
		return strconv.AppendInt([]byte{}, int64(x), 10), nil
	case int32:
		return strconv.AppendInt([]byte{}, int64(x), 10), nil
	case int64:
		return strconv.AppendInt([]byte{}, x, 10), nil
	case uint:
		return strconv.AppendUint([]byte{}, uint64(x), 10), nil
	case uint8:
		return strconv.AppendUint([]byte{}, uint64(x), 10), nil
	case uint16:
		return strconv.AppendUint([]byte{}, uint64(x), 10), nil
	case uint32:
		return strconv.AppendUint([]byte{}, uint64(x), 10), nil
	case uint64:
		return strconv.AppendUint([]byte{}, x, 10), nil
	case float32:
		return strconv.AppendFloat([]byte{}, float64(x), 'g', -1, 32), nil
	case float64:
		return strconv.AppendFloat([]byte{}, x, 'g', -1, 64), nil
	default:
		rv := reflect.ValueOf(x)

		for rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil, nil
			}
			rv = rv.Elem()
		}

		switch rv.Kind() {
		case reflect.String:
			return []byte(rv.String()), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return strconv.AppendInt([]byte{}, rv.Int(), 10), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return strconv.AppendUint([]byte{}, rv.Uint(), 10), nil
		case reflect.Float32:
			return strconv.AppendFloat([]byte{}, rv.Float(), 'g', -1, 32), nil
		case reflect.Float64:
			return strconv.AppendFloat([]byte{}, rv.Float(), 'g', -1, 64), nil
		case reflect.Bool:
			return strconv.AppendBool([]byte{}, rv.Bool()), nil
		}

		return nil, fmt.Errorf("unsupported type %T", x)
	}
}

func UnmarshalText(v interface{}, data []byte) error {
	if rv, ok := v.(reflect.Value); ok {
		if rv.Kind() != reflect.Ptr {
			rv = rv.Addr()
		} else {
			// NewPtrTo ptr value
			if rv.IsNil() {
				rv.Set(New(rv.Type()))
			}
		}

		if textUnmarshaler, ok := rv.Interface().(encoding.TextUnmarshaler); ok {
			if err := textUnmarshaler.UnmarshalText(data); err != nil {
				return pkgerrors.Wrapf(err, "unmarshal text to %T failed", v)
			}
			return nil
		}

		return unmarshalTextToReflectValue(rv, data)
	}

	if textUnmarshaler, ok := v.(encoding.TextUnmarshaler); ok {
		if err := textUnmarshaler.UnmarshalText(data); err != nil {
			return pkgerrors.Wrapf(err, "unmarshal text to %T failed", v)
		}
		return nil
	}

	if v == nil {
		return UnmarshalText(reflect.ValueOf(v), data)
	}

	switch x := v.(type) {
	case *string:
		*x = string(data)
	case *bool:
		v, err := strconv.ParseBool(string(data))
		if err != nil {
			return pkgerrors.Wrapf(err, "unmarshal text")
		}
		*x = v
	case *int:
		i, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = int(i)
	case *int8:
		i, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = int8(i)
	case *int16:
		i, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = int16(i)
	case *int32:
		i, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = int32(i)
	case *int64:
		i, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = i
	case *uint:
		i, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = uint(i)
	case *uint8:
		i, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = uint8(i)
	case *uint16:
		i, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = uint16(i)
	case *uint32:
		i, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = uint32(i)
	case *uint64:
		i, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = i
	case *float32:
		i, err := strconv.ParseFloat(string(data), 32)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = float32(i)
	case *float64:
		i, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		*x = i
	default:
		return unmarshalTextToReflectValue(reflect.ValueOf(x), data)
	}
	return nil
}

func unmarshalTextToReflectValue(rv reflect.Value, data []byte) error {
	if rv.Kind() != reflect.Ptr {
		return pkgerrors.Errorf("unmarshal text need ptr value, but got %#v", rv.Interface())
	}

	// NewPtrTo ptr value
	if rv.IsNil() {
		rv.Set(New(rv.Type()))
	}

	rv = rv.Elem()

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(string(data))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intV, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		rv.SetInt(intV)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintV, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		rv.SetUint(uintV)
	case reflect.Float32, reflect.Float64:
		floatV, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		rv.SetFloat(floatV)
	case reflect.Bool:
		boolV, err := strconv.ParseBool(string(data))
		if err != nil {
			return pkgerrors.Wrap(err, "unmarshal text")
		}
		rv.SetBool(boolV)
	}
	return nil
}
