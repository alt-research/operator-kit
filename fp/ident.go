package fp

// Ident returns the value passed in.
func Ident[T any](v T) T {
	return v
}
