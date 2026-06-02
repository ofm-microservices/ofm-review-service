package cursordb

// Encoder encodes and decodes the canonical DB cursor for review pagination.
type Encoder interface {
	Encode(lastScore int64, lastID string) string
}
