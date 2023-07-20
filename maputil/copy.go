package maputil

func Merge[T comparable, P any](dst *map[T]P, src map[T]P) {
	if len(src) == 0 {
		return
	}
	if dst == nil || *dst == nil {
		*dst = make(map[T]P)
	}
	for k, v := range src {
		if _, ok := (*dst)[k]; !ok {
			(*dst)[k] = v
		}
	}
}
func MergeOverwrite[T comparable, P any](dst *map[T]P, src map[T]P) {
	if len(src) == 0 {
		return
	}
	if dst == nil || *dst == nil {
		*dst = make(map[T]P)
	}
	for k, v := range src {
		(*dst)[k] = v
	}
}
