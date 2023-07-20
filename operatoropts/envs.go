package operatoropts

import (
	"os"
	"strconv"
	"strings"
)

func GetEnv(key, val string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return val
	}
	return v
}

func GetEnvInt(key string, val int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return val, nil
	}
	return strconv.Atoi(v)
}

func GetEnvBool(key string, val bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return val
	}
	switch strings.ToLower(v) {
	case "true", "t", "yes", "y", "on", "1":
		return true
	}
	return false
}
