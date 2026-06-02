package nats

import (
	"errors"
	"testing"
)

func TestAnnotateNATSErrorHelpers(t *testing.T) {
	base := errors.New("boom")
	if got := AnnotateEnsureStreamError("stream", base, base); got == nil {
		t.Fatalf("expected ensure stream annotation")
	}
	if got := AnnotateConnectToNATSError(base); got == nil {
		t.Fatalf("expected connect annotation")
	}
	if got := AnnotateInitJetStreamContextError(base); got == nil {
		t.Fatalf("expected jetstream annotation")
	}
}
