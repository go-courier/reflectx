package reflectx

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-courier/ptr"
	. "github.com/onsi/gomega"
)

type Bytes []byte

func BenchmarkIsBytes(b *testing.B) {
	b.Run("Raw Bytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsBytes([]byte(""))
		}
	})
	b.Run("Named Bytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsBytes(Bytes(""))
		}
	})
}

func TestIsBytes(t *testing.T) {
	t.Run("Raw Bytes", func(t *testing.T) {
		NewWithT(t).Expect(IsBytes([]byte(""))).To(BeTrue())
	})
	t.Run("Named Bytes", func(t *testing.T) {
		NewWithT(t).Expect(IsBytes(Bytes(""))).To(BeTrue())
	})
	t.Run("Others", func(t *testing.T) {
		NewWithT(t).Expect(IsBytes("")).To(BeFalse())
		NewWithT(t).Expect(IsBytes(true)).To(BeFalse())
	})
}

func TestFullTypeName(t *testing.T) {
	NewWithT(t).Expect(FullTypeName(reflect.TypeOf(ptr.Int(1)))).To(Equal("*int"))
	NewWithT(t).Expect(FullTypeName(reflect.PtrTo(reflect.TypeOf(1)))).To(Equal("*int"))
	NewWithT(t).Expect(FullTypeName(reflect.PtrTo(reflect.TypeOf(time.Now())))).To(Equal("*time.Time"))
	NewWithT(t).Expect(FullTypeName(reflect.PtrTo(reflect.TypeOf(struct {
		Name string
	}{})))).To(Equal("*struct { Name string }"))
}

func TestIndirectType(t *testing.T) {
	NewWithT(t).Expect(reflect.TypeOf(1)).To(Equal(Deref(reflect.TypeOf(ptr.Int(1)))))
	NewWithT(t).Expect(reflect.TypeOf(1)).To(Equal(Deref(reflect.PtrTo(reflect.TypeOf(1)))))

	tpe := reflect.TypeOf(1)
	for i := 0; i < 10; i++ {
		tpe = reflect.PtrTo(tpe)
	}
	NewWithT(t).Expect(reflect.TypeOf(1)).To(Equal(Deref(tpe)))
}
