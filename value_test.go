package reflectx

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	. "github.com/onsi/gomega"
)

func TestIndirect(t *testing.T) {
	NewWithT(t).Expect(reflect.ValueOf(1).Interface()).To(Equal(Indirect(reflect.ValueOf(ptr.Int(1))).Interface()))
	NewWithT(t).Expect(reflect.ValueOf(0).Interface()).To(Equal(Indirect(reflect.New(reflect.TypeOf(0))).Interface()))

	rv := New(reflect.PtrTo(reflect.PtrTo(reflect.PtrTo(reflect.TypeOf("")))))
	NewWithT(t).Expect(reflect.ValueOf("").Interface()).To(Equal(Indirect(rv).Interface()))
}

type Zero string

func (Zero) IsZero() bool {
	return true
}

func BenchmarkNew(b *testing.B) {
	tpe := reflect.PtrTo(reflect.TypeOf(Zero("")))

	for i := 0; i < b.N; i++ {
		_ = New(tpe)
	}
}

func TestNew(t *testing.T) {
	t.Run("NewType", func(t *testing.T) {
		tpe := reflect.TypeOf(Zero(""))
		_, ok := New(tpe).Interface().(Zero)
		NewWithT(t).Expect(ok).To(BeTrue())
	})

	t.Run("NewPtrType", func(t *testing.T) {
		tpe := reflect.PtrTo(reflect.TypeOf(Zero("")))
		_, ok := New(tpe).Interface().(*Zero)
		NewWithT(t).Expect(ok).To(BeTrue())
	})

	t.Run("NewPtrPtrType", func(t *testing.T) {
		tpe := reflect.PtrTo(reflect.PtrTo(reflect.TypeOf(Zero(""))))
		_, ok := New(tpe).Interface().(**Zero)
		NewWithT(t).Expect(ok).To(BeTrue())
	})
}

type S struct {
	V interface{}
}

var emptyValues = []interface{}{
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

var nonEmptyValues = []interface{}{
	Zero("11111111111"),
	ptr.String("12322"),
}

func BenchmarkIsEmptyValue(b *testing.B) {
	for i, v := range append(emptyValues, nonEmptyValues...) {
		b.Run(fmt.Sprintf("%d: %#v", i, v), func(b *testing.B) {
			IsEmptyValue(v)
		})

		if _, ok := v.(reflect.Value); !ok {
			rv := reflect.ValueOf(v)
			b.Run(fmt.Sprintf("%d: reflect.Value(%#v)", i, v), func(b *testing.B) {
				IsEmptyValue(rv)
			})
		}
	}
}

func TestIsEmptyValue(t *testing.T) {
	for i, v := range emptyValues {
		t.Run(fmt.Sprintf("%d: %#v", i, v), func(t *testing.T) {
			NewWithT(t).Expect(IsEmptyValue(v)).To(BeTrue())
		})

		if _, ok := v.(reflect.Value); !ok {
			rv := reflect.ValueOf(v)

			t.Run(fmt.Sprintf("%d: reflect.Value(%#v)", i, v), func(t *testing.T) {
				NewWithT(t).Expect(IsEmptyValue(rv)).To(BeTrue())
			})
		}
	}
}
