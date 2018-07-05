package reflectx

import (
	"reflect"
	"bytes"
)

func FullTypeName(tpe reflect.Type) string {
	buf := bytes.NewBuffer(nil)

	for tpe.Kind() == reflect.Ptr {
		buf.WriteByte('*')
		tpe = tpe.Elem()
	}

	if pkgPath := tpe.PkgPath(); pkgPath != "" {
		buf.WriteString(pkgPath)
		buf.WriteRune('.')
	}

	buf.WriteString(tpe.Name())
	return buf.String()
}

func IndirectType(tpe reflect.Type) reflect.Type {
	if tpe.Kind() == reflect.Ptr {
		return IndirectType(tpe.Elem())
	}
	return tpe
}
