package typesutil

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var (
	typesCache = sync.Map{}
	pkgCache   = sync.Map{}
)

func NewPackage(importPath string) *types.Package {
	if v, ok := pkgCache.Load(importPath); ok {
		return v.(*types.Package)
	}
	pkg, err := importer.ForCompiler(token.NewFileSet(), "source", nil).Import(importPath)
	if err != nil && importPath != "" {
		panic(err)
	}
	pkgCache.Store(importPath, pkg)
	return pkg
}

func TypeByName(importPath string, name string) types.Type {
	pkg := NewPackage(importPath)
	if pkg == nil {
		return nil
	}
	return pkg.Scope().Lookup(name).Type()
}

func NewTypesTypeFromReflectType(rtype reflect.Type) types.Type {
	underlying := func() types.Type {
		switch rtype.Kind() {
		case reflect.Array:
			return types.NewArray(NewTypesTypeFromReflectType(rtype.Elem()), int64(rtype.Len()))
		case reflect.Slice:
			return types.NewSlice(NewTypesTypeFromReflectType(rtype.Elem()))
		case reflect.Map:
			return types.NewMap(NewTypesTypeFromReflectType(rtype.Key()), NewTypesTypeFromReflectType(rtype.Elem()))
		case reflect.Chan:
			return types.NewChan(types.ChanDir(rtype.ChanDir()), NewTypesTypeFromReflectType(rtype.Elem()))
		case reflect.Func:
			params := make([]*types.Var, rtype.NumIn())
			for i := range params {
				param := rtype.In(i)
				params[i] = types.NewParam(0, NewPackage(param.PkgPath()), "", NewTypesTypeFromReflectType(param))
			}
			results := make([]*types.Var, rtype.NumOut())
			for i := range results {
				result := rtype.Out(i)
				results[i] = types.NewParam(0, NewPackage(result.PkgPath()), "", NewTypesTypeFromReflectType(result))
			}
			return types.NewSignature(
				nil,
				types.NewTuple(params...),
				types.NewTuple(results...),
				rtype.IsVariadic(),
			)
		case reflect.Interface:
			funcs := make([]*types.Func, rtype.NumMethod())
			for i := range funcs {
				m := rtype.Method(i)

				funcs[i] = types.NewFunc(
					0,
					NewPackage(m.PkgPath),
					m.Name,
					NewTypesTypeFromReflectType(m.Type).(*types.Signature),
				)
			}
			return types.NewInterfaceType(funcs, nil).Complete()
		case reflect.Struct:
			fields := make([]*types.Var, rtype.NumField())
			tags := make([]string, len(fields))
			for i := range fields {
				f := rtype.Field(i)
				fields[i] = types.NewField(
					0,
					NewPackage(f.PkgPath),
					f.Name,
					NewTypesTypeFromReflectType(f.Type),
					f.Anonymous,
				)
				tags[i] = string(f.Tag)
			}
			return types.NewStruct(fields, tags)
		case reflect.Bool:
			return types.Typ[types.Bool]
		case reflect.Int:
			return types.Typ[types.Int]
		case reflect.Int8:
			return types.Typ[types.Int8]
		case reflect.Int16:
			return types.Typ[types.Int16]
		case reflect.Int32:
			return types.Typ[types.Int32]
		case reflect.Int64:
			return types.Typ[types.Int64]
		case reflect.Uint:
			return types.Typ[types.Uint]
		case reflect.Uint8:
			return types.Typ[types.Uint8]
		case reflect.Uint16:
			return types.Typ[types.Uint16]
		case reflect.Uint32:
			return types.Typ[types.Uint32]
		case reflect.Uint64:
			return types.Typ[types.Uint64]
		case reflect.Uintptr:
			return types.Typ[types.Uintptr]
		case reflect.Float32:
			return types.Typ[types.Float32]
		case reflect.Float64:
			return types.Typ[types.Float64]
		case reflect.Complex64:
			return types.Typ[types.Complex64]
		case reflect.Complex128:
			return types.Typ[types.Complex128]
		case reflect.String:
			return types.Typ[types.String]
		case reflect.UnsafePointer:
			return types.Typ[types.UnsafePointer]
		}
		return nil
	}

	ptrCount := 0

	mayWithPtr := func(typ types.Type) types.Type {
		for ptrCount > 0 {
			typ = types.NewPointer(typ)
			ptrCount--
		}
		return typ
	}

	for rtype.Kind() == reflect.Ptr {
		rtype = rtype.Elem()
		ptrCount++
	}

	name := rtype.Name()
	pkgPath := rtype.PkgPath()

	if name == "error" && pkgPath == "" {
		return mayWithPtr(TypeByName("errors", "New").Underlying().(*types.Signature).Results().At(0).Type())
	}

	if pkgPath != "" {
		key := name
		if pkgPath != "" {
			key = pkgPath + "." + name
		}

		if typ, ok := typesCache.Load(key); ok {
			return mayWithPtr(typ.(types.Type))
		}

		ttype := TypeByName(pkgPath, name)
		typesCache.Store(key, ttype)
		return mayWithPtr(ttype)
	}

	return mayWithPtr(underlying())
}

func FromTType(ttype types.Type) *TType {
	return &TType{
		Type: ttype,
	}
}

type TType struct {
	Type types.Type

	methodScanned bool
	methods       []*types.Func
	ptrMethods    []*types.Func
}

func methodsOf(typ types.Type) (methods []*TMethod) {
	switch t := typ.(type) {
	case *types.Named:
		for i := 0; i < t.NumMethods(); i++ {
			methodInfo := &TMethod{
				Func: t.Method(i),
			}
			if _, ok := methodInfo.Func.Type().(*types.Signature).Recv().Type().(*types.Pointer); ok {
				methodInfo.Ptr = true
			}
			methods = append(methods, methodInfo)
		}

		methods = append(methods, methodsOf(t.Underlying())...)
	case *types.Pointer:
		methods = append(methods, methodsOf(t.Elem())...)
	case *types.Struct:
		for i := 0; i < t.NumFields(); i++ {
			field := t.Field(i)
			if field.Anonymous() {
				methods = append(methods, methodsOf(field.Type())...)
			}
		}
	}
	return
}

func (ttype *TType) tryScanMethods() {
	if !ttype.methodScanned {
		ttype.methodScanned = true
		methods := methodsOf(ttype.Type)
		for i := range methods {
			m := methods[i]
			if !m.Ptr {
				ttype.methods = append(ttype.methods, m.Func)
			}
			ttype.ptrMethods = append(ttype.ptrMethods, m.Func)
		}
	}
}

func (ttype *TType) NumMethod() int {
	if ttype.Kind() == reflect.Interface {
		switch t := ttype.Type.(type) {
		case *types.Named:
			return t.Underlying().(*types.Interface).NumMethods()
		case *types.Interface:
			return t.NumMethods()
		}
	}

	ttype.tryScanMethods()

	switch ttype.Type.(type) {
	case *types.Pointer:
		return len(ttype.ptrMethods)
	default:
		return len(ttype.methods)
	}
}

func (ttype *TType) Method(i int) Method {
	if ttype.Kind() == reflect.Interface {
		switch t := ttype.Type.(type) {
		case *types.Named:
			return &TMethod{Recv: ttype, Func: t.Underlying().(*types.Interface).Method(i)}
		case *types.Interface:
			return &TMethod{Recv: ttype, Func: t.Method(i)}
		}
	}

	ttype.tryScanMethods()

	switch ttype.Type.(type) {
	case *types.Pointer:
		if ttype.ptrMethods != nil {
			return &TMethod{Recv: ttype, Func: ttype.ptrMethods[i]}
		}
	default:
		if ttype.methods != nil {
			return &TMethod{Recv: ttype, Func: ttype.methods[i]}
		}
	}

	return nil
}

func (ttype *TType) MethodByName(name string) (Method, bool) {
	for i := 0; i < ttype.NumMethod(); i++ {
		f := ttype.Method(i)
		if f.Name() == name {
			return f, true
		}
	}
	return nil, false
}

func (ttype *TType) NumIn() int {
	switch t := ttype.Type.(type) {
	case *types.Named:
		return FromTType(t.Underlying()).NumIn()
	case *types.Signature:
		return t.Params().Len()
	}
	return 0
}

func (ttype *TType) In(i int) Type {
	switch t := ttype.Type.(type) {
	case *types.Named:
		return FromTType(t.Underlying()).In(i)
	case *types.Signature:
		return FromTType(t.Params().At(i).Type())
	}
	return nil
}

func (ttype *TType) NumOut() int {
	switch t := ttype.Type.(type) {
	case *types.Named:
		return FromTType(t.Underlying()).NumOut()
	case *types.Signature:
		return t.Results().Len()
	}
	return 0
}

func (ttype *TType) Out(i int) Type {
	switch t := ttype.Type.(type) {
	case *types.Named:
		return FromTType(t.Underlying()).Out(i)
	case *types.Signature:
		return FromTType(t.Results().At(i).Type())
	}
	return nil
}

func (ttype *TType) Implements(u Type) bool {
	return types.Implements(ttype.Type, u.(*TType).Type.(*types.Interface))
}

func (ttype *TType) AssignableTo(u Type) bool {
	return types.AssignableTo(ttype.Type, u.(*TType).Type)
}

func (ttype *TType) ConvertibleTo(u Type) bool {
	return types.ConvertibleTo(ttype.Type, u.(*TType).Type)
}

func (ttype *TType) Comparable() bool {
	return types.Comparable(ttype.Type)
}

func (ttype *TType) Field(i int) StructField {
	switch t := ttype.Type.(type) {
	case *types.Named:
		return FromTType(t.Underlying()).Field(i)
	case *types.Struct:
		return &TStructField{Var: t.Field(i), TagStr: t.Tag(i)}
	}
	return nil
}

func (ttype *TType) FieldByName(name string) (StructField, bool) {
	return ttype.FieldByNameFunc(func(s string) bool {
		return name == s
	})
}

func (ttype *TType) FieldByNameFunc(match func(string) bool) (StructField, bool) {
	for i := 0; i < ttype.NumField(); i++ {
		f := ttype.Field(i)
		if match(f.Name()) {
			return f, true
		}
		if f.Anonymous() {
			if sf, ok := f.Type().FieldByNameFunc(match); ok {
				return sf, ok
			}
		}
	}
	return nil, false
}

func (ttype *TType) NumField() int {
	switch t := ttype.Type.(type) {
	case *types.Pointer:
		return FromTType(t.Elem()).NumField()
	case *types.Named:
		return FromTType(t.Underlying()).NumField()
	case *types.Struct:
		return t.NumFields()
	}
	return 0
}

func (ttype *TType) Kind() reflect.Kind {
	switch t := ttype.Type.(type) {
	case *types.Named:
		pkg := t.Obj().Pkg()
		if pkg != nil && pkg.Name() == "unsafe" && t.Obj().Name() == "Pointer" {
			return reflect.UnsafePointer
		}
		return FromTType(t.Underlying()).Kind()
	case *types.Interface:
		return reflect.Interface
	case *types.Pointer:
		return reflect.Ptr
	case *types.Struct:
		return reflect.Struct
	case *types.Map:
		return reflect.Map
	case *types.Slice:
		return reflect.Slice
	case *types.Array:
		return reflect.Array
	case *types.Chan:
		return reflect.Chan
	case *types.Signature:
		return reflect.Func
	case *types.Basic:
		switch t.Kind() {
		case types.UntypedBool, types.Bool:
			return reflect.Bool
		case types.UntypedInt, types.Int:
			return reflect.Int
		case types.Int8:
			return reflect.Int8
		case types.Int16:
			return reflect.Int16
		case types.Int32, types.UntypedRune:
			// includes types.Rune
			return reflect.Int32
		case types.Int64:
			return reflect.Int64
		case types.Uint:
			return reflect.Uint
		case types.Uint8:
			// includes types.Byte
			return reflect.Uint8
		case types.Uint16:
			return reflect.Uint16
		case types.Uint32:
			return reflect.Uint32
		case types.Uint64:
			return reflect.Uint64
		case types.Uintptr:
			return reflect.Uintptr
		case types.Float32, types.UntypedFloat:
			return reflect.Float32
		case types.Float64:
			return reflect.Float64
		case types.Complex64, types.UntypedComplex:
			return reflect.Complex64
		case types.Complex128:
			return reflect.Complex128
		case types.String, types.UntypedString:
			return reflect.String
		case types.UnsafePointer:
			return reflect.UnsafePointer
		}
	}
	return reflect.Invalid
}

func (ttype *TType) Name() string {
	switch t := ttype.Type.(type) {
	case *types.Named:
		return t.Obj().Name()
	case *types.Basic:
		return t.Name()
	}
	return ""
}

func (ttype *TType) PkgPath() string {
	switch x := ttype.Type.(type) {
	case *types.Named:
		return x.Obj().Pkg().Path()
	case *types.Basic:
		// unsafe.Pointer as basic since 1.17
		if strings.HasPrefix(x.String(), "unsafe.") {
			return "unsafe"
		}
	}
	return ""
}

func (ttype *TType) Key() Type {
	if named, ok := ttype.Type.(*types.Named); ok {
		return FromTType(named.Underlying()).Key()
	}
	if typ, ok := ttype.Type.(interface{ Key() types.Type }); ok {
		return FromTType(typ.Key())
	}
	return nil
}

func (ttype *TType) Elem() Type {
	if named, ok := ttype.Type.(*types.Named); ok {
		return FromTType(named.Underlying()).Elem()
	}
	if typ, ok := ttype.Type.(interface{ Elem() types.Type }); ok {
		return FromTType(typ.Elem())
	}
	return nil
}

func (ttype *TType) Len() int {
	switch typ := ttype.Type.(type) {
	case *types.Named:
		return FromTType(typ.Underlying()).Len()
	case *types.Array:
		return int(typ.Len())
	}
	return 0
}

func (ttype *TType) String() string {
	typeString := func(typ types.Type) string {
		return types.TypeString(typ, func(pkg *types.Package) string {
			return pkg.Name()
		})
	}

	switch t := ttype.Type.(type) {
	case *types.Basic:
		return ttype.Kind().String()
	case *types.Slice:
		return "[]" + FromTType(t.Elem()).String()
	case *types.Array:
		return fmt.Sprintf("[%d]", t.Len()) + FromTType(t.Elem()).String()
	case *types.Map:
		return fmt.Sprintf("map[%s]%s", FromTType(t.Key()), FromTType(t.Elem()))
	case *types.Chan:
		return "chan " + FromTType(t.Elem()).String()
	case *types.Struct:
		buf := bytes.NewBuffer(nil)
		buf.WriteString("struct {")
		n := t.NumFields()
		for i := 0; i < n; i++ {
			buf.WriteRune(' ')
			f := t.Field(i)
			if !f.Anonymous() {
				buf.WriteString(f.Name())
				buf.WriteRune(' ')
			}
			buf.WriteString(FromTType(f.Type()).String())

			tag := t.Tag(i)
			if tag != "" {
				buf.WriteRune(' ')
				buf.WriteString(strconv.Quote(tag))
			}

			if i == n-1 {
				buf.WriteRune(' ')
			} else {
				buf.WriteRune(';')
			}
		}
		buf.WriteString("}")
		return buf.String()
	case *types.Interface:
		buf := bytes.NewBuffer(nil)
		buf.WriteString("interface {")
		n := t.NumMethods()
		for i := 0; i < n; i++ {
			buf.WriteRune(' ')
			m := &TMethod{Func: t.Method(i)}

			pkgPath := m.PkgPath()
			if pkgPath != "" {
				pkg := NewPackage(pkgPath)
				buf.WriteString(pkg.Name())
				buf.WriteRune('.')
			}

			buf.WriteString(m.Name())
			buf.WriteString(m.Type().String()[4:])

			if i == n-1 {
				buf.WriteRune(' ')
			} else {
				buf.WriteRune(';')
			}
		}
		buf.WriteString("}")
		return buf.String()
	case *types.Signature:
		buf := bytes.NewBuffer(nil)
		buf.WriteString("func(")
		{
			params := t.Params()
			n := params.Len()

			recv := t.Recv()
			if recv != nil {
				switch recvTyp := recv.Type().(type) {
				case *types.Pointer:
					elem := recvTyp.Elem()

					if FromTType(elem).Kind() != reflect.Interface {
						buf.WriteRune('*')
						buf.WriteString(typeString(elem))
						if n > 0 {
							buf.WriteString(", ")
						}
					}
				case *types.Named:
					if FromTType(recvTyp).Kind() != reflect.Interface {
						buf.WriteString(typeString(recvTyp))
						if n > 0 {
							buf.WriteString(", ")
						}
					}
				case *types.Struct:
					buf.WriteString(FromTType(recvTyp).String())
					if n > 0 {
						buf.WriteString(", ")
					}
				}
			}

			for i := 0; i < n; i++ {
				p := params.At(i)

				if i == n-1 && t.Variadic() {
					buf.WriteString("...")
					buf.WriteString(FromTType(p.Type().(*types.Slice).Elem()).String())
				} else {
					buf.WriteString(FromTType(p.Type()).String())
				}

				if i < n-1 {
					buf.WriteString(", ")
				}
			}
			buf.WriteString(")")
		}

		{
			results := t.Results()
			n := results.Len()
			if n > 0 {
				buf.WriteRune(' ')
			}
			if n > 1 {
				buf.WriteString("(")
			}
			for i := 0; i < n; i++ {
				if i > 0 {
					buf.WriteString(", ")
				}

				r := results.At(i)
				buf.WriteString(FromTType(r.Type()).String())
			}
			if n > 1 {
				buf.WriteString(")")
			}
		}

		return buf.String()
	}

	return typeString(ttype.Type)
}

type TStructField struct {
	*types.Var
	TagStr string
}

func (f *TStructField) PkgPath() string {
	if ast.IsExported(f.Name()) {
		return ""
	}
	pkg := f.Var.Pkg()
	if pkg != nil {
		return pkg.Path()
	}
	return ""
}

func (f *TStructField) Tag() reflect.StructTag {
	return reflect.StructTag(f.TagStr)
}

func (f *TStructField) Type() Type {
	return FromTType(f.Var.Type())
}

type TMethod struct {
	Ptr  bool
	Recv *TType
	Func *types.Func
}

func (m *TMethod) PkgPath() string {
	if ast.IsExported(m.Name()) {
		return ""
	}
	pkg := m.Func.Pkg()
	if pkg != nil {
		return pkg.Path()
	}
	return ""
}

func (m *TMethod) Name() string {
	return m.Func.Name()
}

func (m *TMethod) Type() Type {
	s := m.Func.Type().(*types.Signature)
	if m.Recv == nil {
		return FromTType(s)
	}

	pkg := (*types.Package)(nil)
	if named, ok := m.Recv.Type.(*types.Named); ok {
		pkg = named.Obj().Pkg()
	}

	return FromTType(types.NewSignature(
		types.NewVar(0, pkg, "", m.Recv.Type),
		s.Params(),
		s.Results(),
		s.Variadic(),
	))
}
