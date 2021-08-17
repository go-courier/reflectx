package reflectx

type ZeroChecker interface {
	IsZero() bool
}

type RawValuer interface {
	Set(v interface{}) error
}
