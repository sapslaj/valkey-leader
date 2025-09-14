// utility functions for dealing with pointers aka Go's `Optional<T>`.
package ptr

func Of[T any](v T) *T {
	return &v
}

func FromDefault[T any](v *T, def T) T {
	if v != nil {
		return *v
	}
	return def
}

func From[T any](v *T) T {
	var def T
	return FromDefault(v, def)
}

func SlicesOf[T any](v []T) []*T {
	s := make([]*T, len(v))
	for i := range v {
		s[i] = &v[i]
	}
	return s
}

func SlicesFrom[T any](v []*T) []T {
	s := make([]T, len(v))
	for i := range v {
		s[i] = *v[i]
	}
	return s
}
