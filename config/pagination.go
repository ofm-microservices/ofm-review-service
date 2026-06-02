package config

import "time"

// ReviewPaginationConfig controls the fixed page/window sizing used by the
// review read path and encrypted cursor tokens.
type ReviewPaginationConfig struct {
	PageSize     int           `env:"REVIEW_PAGE_SIZE" envDefault:"10"`
	WindowSize   int           `env:"REVIEW_WINDOW_SIZE" envDefault:"100"`
	WindowTTL    time.Duration `env:"REVIEW_WINDOW_TTL" envDefault:"15m"`
	CursorSecret string        `env:"REVIEW_CURSOR_SECRET,required"`
}
