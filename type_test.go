package reflectx

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"
)

type Bytes []byte

func TestIsBytes(t *testing.T) {
	assert.True(t, IsBytes(reflect.TypeOf([]byte(""))))
	assert.True(t, IsBytes(reflect.TypeOf(Bytes(""))))
	assert.False(t, IsBytes(reflect.TypeOf("")))
	assert.False(t, IsBytes(reflect.TypeOf(true)))
}

func TestFullTypeName(t *testing.T) {
	assert.Equal(t, "*int", FullTypeName(reflect.TypeOf(ptr.Int(1))))
	assert.Equal(t, "*int", FullTypeName(reflect.PtrTo(reflect.TypeOf(1))))
	assert.Equal(t, "*(time)time.Time", FullTypeName(reflect.PtrTo(reflect.TypeOf(time.Now()))))
	assert.Equal(t, "*struct { Name string }", FullTypeName(reflect.PtrTo(reflect.TypeOf(struct {
		Name string
	}{}))))
}

func TestIndirectType(t *testing.T) {
	assert.Equal(t, IndirectType(reflect.TypeOf(ptr.Int(1))), reflect.TypeOf(1))
	assert.Equal(t, IndirectType(reflect.PtrTo(reflect.TypeOf(1))), reflect.TypeOf(1))

	tpe := reflect.TypeOf(1)
	for i := 0; i < 10; i++ {
		tpe = reflect.PtrTo(tpe)
	}
	assert.Equal(t, IndirectType(tpe), reflect.TypeOf(1))
}
