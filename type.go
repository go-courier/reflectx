package reflectx

import (
	"bytes"
	"reflect"
)

func IsBytes(v interface{}) bool {
	switch x := v.(type) {
	case []byte:
		return true
	case reflect.Type:
		return x.Kind() == reflect.Slice && x.Elem().Kind() == reflect.Uint8
	default:
		return IsBytes(reflect.TypeOf(v))
	}
}

func FullTypeName(rtype reflect.Type) string {
	buf := bytes.NewBuffer(nil)

	for rtype.Kind() == reflect.Ptr {
		buf.WriteByte('*')
		rtype = rtype.Elem()
	}

	if name := rtype.Name(); name != "" {
		if pkgPath := rtype.PkgPath(); pkgPath != "" {
			buf.WriteString(pkgPath)
			buf.WriteRune('.')
		}
		buf.WriteString(name)
		return buf.String()
	}

	buf.WriteString(rtype.String())
	return buf.String()
}

func Deref(tpe reflect.Type) reflect.Type {
	if tpe.Kind() == reflect.Ptr {
		return Deref(tpe.Elem())
	}
	return tpe
}
