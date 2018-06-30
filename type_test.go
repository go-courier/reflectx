package reflectx

import (
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"
)

func TestIndirectType(t *testing.T) {
	assert.Equal(t, IndirectType(reflect.TypeOf(ptr.Int(1))), reflect.TypeOf(1))
	assert.Equal(t, IndirectType(reflect.PtrTo(reflect.TypeOf(1))), reflect.TypeOf(1))

	tpe := reflect.TypeOf(1)
	for i := 0; i < 10; i++ {
		tpe = reflect.PtrTo(tpe)
	}
	assert.Equal(t, IndirectType(tpe), reflect.TypeOf(1))
}
