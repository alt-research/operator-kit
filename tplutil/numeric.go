package tplutil

import "github.com/spf13/cast"

func incr(i interface{}) int64 {
	return cast.ToInt64(i) + 1
}

func init() {
	FuncMap["incr"] = incr
}
