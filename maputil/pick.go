package maputil

func Pick[T comparable, P any](src map[T]P, picker func(T, P) bool) map[T]P {
	dst := make(map[T]P)
	if len(src) == 0 {
		return dst
	}
	for k, v := range src {
		if picker(k, v) {
			dst[k] = v
		}
	}
	return dst
}
