package reflectx

import (
	"reflect"
)

func Indirect(rv reflect.Value) reflect.Value {
	if rv.Kind() == reflect.Ptr {
		return Indirect(rv.Elem())
	}
	return rv
}

func New(tpe reflect.Type) reflect.Value {
	rv := reflect.New(tpe).Elem()
	if tpe.Kind() == reflect.Ptr {
		rv.Set(New(tpe.Elem()).Addr())
		return rv
	}
	return rv
}

func IsEmptyValue(v interface{}) bool {
	if rv, ok := v.(reflect.Value); ok {
		if rv.Kind() == reflect.Ptr && rv.IsNil() {
			return true
		}

		if rv.IsValid() && rv.CanInterface() {
			if zeroChecker, ok := rv.Interface().(ZeroChecker); ok {
				return zeroChecker.IsZero()
			}
		}

		return isEmptyReflectValue(rv)
	}

	if zeroChecker, ok := v.(ZeroChecker); ok {
		return zeroChecker.IsZero()
	}

	switch x := v.(type) {
	case string:
		return x == ""
	case bool:
		return !x
	case int:
		return x == 0
	case int8:
		return x == 0
	case int16:
		return x == 0
	case int32:
		return x == 0
	case int64:
		return x == 0
	case uint:
		return x == 0
	case uint8:
		return x == 0
	case uint16:
		return x == 0
	case uint32:
		return x == 0
	case uint64:
		return x == 0
	case float32:
		return x == 0
	case float64:
		return x == 0
	case []byte:
		return len(x) == 0
	case []interface{}:
		return len(x) == 0
	default:
		return isEmptyReflectValue(reflect.ValueOf(x))
	}
}

func isEmptyReflectValue(rv reflect.Value) bool {
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return true
	}

	if rv.IsValid() && rv.CanInterface() {
		if zeroChecker, ok := rv.Interface().(ZeroChecker); ok {
			return zeroChecker.IsZero()
		}
	}

	switch rv.Kind() {
	case reflect.Interface:
		if rv.IsNil() {
			return true
		}
		return IsEmptyValue(rv.Elem())
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Ptr:
		return rv.IsNil()
	case reflect.Invalid:
		return true
	}
	return false
}
