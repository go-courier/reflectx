package reflectx

import (
	"reflect"
)

func IndirectType(tpe reflect.Type) reflect.Type {
	if tpe.Kind() == reflect.Ptr {
		return IndirectType(tpe.Elem())
	}
	return tpe
}
