package cursordb

import (
	"encoding/base64"
	"fmt"
)

type encoder struct{}

// New constructs the canonical DB cursor encoder.
func New() Encoder {
	return encoder{}
}

func (encoder) Encode(lastScore int64, lastID string) string {
	raw := fmt.Sprintf("%d:%s", lastScore, lastID)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}
