package hextool

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func JsonChecksum(d any) string {
	b, _ := json.Marshal(d)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}
