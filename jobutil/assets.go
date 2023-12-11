package jobutil

import (
	"embed"
)

//go:embed scripts/*.sh
var assets embed.FS
