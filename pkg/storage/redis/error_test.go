package redis

import (
	"errors"
	"testing"
)

func TestAnnotateRedisPingError(t *testing.T) {
	if got := AnnotateRedisPingError(errors.New("boom")); got == nil {
		t.Fatalf("expected annotated error")
	}
}
