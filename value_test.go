package reflectx

import (
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"
)

func TestIndirect(t *testing.T) {
	assert.Equal(t, Indirect(reflect.ValueOf(ptr.Int(1))).Interface(), reflect.ValueOf(1).Interface())
	assert.Equal(t, Indirect(reflect.New(reflect.TypeOf(0))).Interface(), reflect.ValueOf(0).Interface())

	rv := New(reflect.PtrTo(reflect.PtrTo(reflect.PtrTo(reflect.TypeOf("")))))
	assert.Equal(t, Indirect(rv).Interface(), reflect.ValueOf("").Interface())
}

type Zero string

func (Zero) IsZero() bool {
	return true
}

func TestIsEmptyValue(t *testing.T) {
	type S struct {
		V interface{}
	}

	emptyValues := []interface{}{
		Zero(""),
		(*string)(nil),
		(interface{})(nil),
		(S{}).V,
		"",
		0,
		uint(0),
		float32(0),
		false,
		reflect.ValueOf(S{}).FieldByName("V"),
		nil,
	}
	for _, v := range emptyValues {
		if rv, ok := v.(reflect.Value); ok {
			assert.True(t, IsEmptyValue(rv))
		} else {
			assert.True(t, IsEmptyValue(v))
			assert.True(t, IsEmptyValue(reflect.ValueOf(v)))
		}

	}
}
