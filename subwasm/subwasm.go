package subwasm

import (
	"os/exec"
)

func Exec(args []string) ([]byte, error) {
	return exec.Command("subwasm", args...).Output()
}

func Info(file string) ([]byte, error) {
	return Exec([]string{"info", "--json", file})
}
