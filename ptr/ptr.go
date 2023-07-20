package ptr

func Of[T any](v T) *T {
	return &v
}

func Slice[T any](vs []T) []*T {
	ps := make([]*T, len(vs))
	for i, v := range vs {
		vv := v
		ps[i] = &vv
	}
	return ps
}

func Map[T comparable, P any](vs map[T]P) map[T]*P {
	ps := make(map[T]*P, len(vs))
	for k, v := range vs {
		vv := v
		ps[k] = &vv
	}
	return ps
}

func To[T any](v *T) T {
	return *v
}

func ToSlice[T any](vs []*T) []T {
	ps := make([]T, len(vs))
	for i, v := range vs {
		vv := v
		ps[i] = *vv
	}
	return ps
}

func ToMap[T comparable, P any](vs map[T]*P) map[T]P {
	ps := make(map[T]P, len(vs))
	for k, v := range vs {
		vv := v
		ps[k] = *vv
	}
	return ps
}
