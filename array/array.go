// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package array

import (
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/alt-research/operator-kit/must"
	"golang.org/x/exp/constraints"
)

func Contains[T any](array []T, target T) bool {
	for _, value := range array {
		if reflect.DeepEqual(value, target) {
			return true
		}
	}
	return false
}

func Removes[T any](array *[]T, target T, count ...int) int {
	if array == nil {
		return 0
	}
	c := len(*array)
	removed := 0
	if len(count) > 0 {
		c = count[0]
	}
	for i, value := range *array {
		if c == 0 {
			break
		}
		if reflect.DeepEqual(value, target) {
			*array = append((*array)[:i], (*array)[i+1:]...)
			c--
			removed++
		}
	}
	return removed
}

func Remove[T any](array *[]T, target T) int {
	if array == nil {
		return 0
	}
	for i, value := range *array {
		if reflect.DeepEqual(value, target) {
			*array = append((*array)[:i], (*array)[i+1:]...)
			return i
		}
	}
	return -1
}

func Count[T any](array []T, f func(v T) bool) int {
	var count int
	for _, v := range array {
		if f(v) {
			count++
		}
	}
	return count
}

func Filter[T any](array []T, f func(v T) bool) []T {
	var out []T
	for _, v := range array {
		if f(v) {
			out = append(out, v)
		}
	}
	return out
}

func Map[T any, P any](array []T, f func(v T) P) []P {
	var out []P
	for _, v := range array {
		out = append(out, f(v))
	}
	return out
}

func Reduce[T any, P any](array []T, f func(v T, o P) P) P {
	var out P
	for _, v := range array {
		out = f(v, out)
	}
	return out
}

func Flatten[T any](array [][]T) []T {
	var out []T
	for _, v := range array {
		out = append(out, v...)
	}
	return out
}

func All(array []bool) bool {
	for _, v := range array {
		if !v {
			return false
		}
	}
	return true
}

func Any(array []bool) bool {
	for _, v := range array {
		if v {
			return true
		}
	}
	return false
}

func Fill[T any](array []T, value T) {
	for i := range array {
		array[i] = value
	}
}

func Gen[T any](n int, f func(i int) T) []T {
	var out []T
	for i := 0; i < n; i++ {
		out = append(out, f(i))
	}
	return out
}

func GenRepeated[T any](n int, value T) []T {
	return Gen(n, func(_ int) T { return value })
}

// Slice python style array slicing
// https://stackoverflow.com/questions/509211/understanding-slice-notation
// slicer: "start:end" or "start:" or ":end" or ":"
// examples:
//
//	Slice([]int{1,2,3,4,5}, "1:3") -> []int{2,3}
//	Slice([]int{1,2,3,4,5}, "1:") -> []int{2,3,4,5}
//	Slice([]int{1,2,3,4,5}, ":-1") -> []int{1,2,3,4}
//	Slice([]int{1,2,3,4,5}, "-2:") -> []int{4,5}
//
// TODO: add step
func Slice[T any](a []T, slicer string) []T {
	split := strings.Split(slicer, ":")
	if len(a) == 0 || len(split) == 0 {
		return a
	}
	start := 0
	end := len(a)
	if len(split) > 0 && split[0] != "" {
		start = must.Two(strconv.Atoi(split[0]))
		if start < 0 {
			start = len(a) + start
		}
	}
	if len(split) > 1 && split[1] != "" {
		end = must.Two(strconv.Atoi(split[1]))
		if end < 0 {
			end = len(a) + end
		}
	}
	if end > len(a) {
		end = len(a)
	}
	return a[start:end]
}

func SliceStr(a string, slicer string) string {
	return string(Slice([]rune(a), slicer))
}

func Casts[T any, P any](a []P) []T {
	var out []T
	for _, v := range a {
		i := reflect.ValueOf(v).Interface()
		out = append(out, i.(T))
	}
	return out
}

func Unique[T comparable](a []T) []T {
	set := make(map[T]struct{})
	var out []T
	for _, v := range a {
		if _, ok := set[v]; !ok {
			out = append(out, v)
			set[v] = struct{}{}
		}
	}
	return out
}

func Ordered[T constraints.Ordered](a []T, reverse bool) []T {
	out := make([]T, len(a))
	copy(out, a)
	if reverse {
		sort.Slice(out, func(i, j int) bool { return out[i] > out[j] })
	} else {
		sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	}
	return out
}

func Ascending[T constraints.Ordered](a []T) []T {
	return Ordered(a, false)
}

func Descending[T constraints.Ordered](a []T) []T {
	return Ordered(a, true)
}
