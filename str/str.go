package str

import "strings"

func EqualFold[T, P ~string](a T, b P) bool {
	return strings.EqualFold(string(a), string(b))
}
