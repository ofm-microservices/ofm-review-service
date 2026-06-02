package nats

import (
	"errors"
	"testing"
)

func TestAnnotateNATSPresentationErrors(t *testing.T) {
	base := errors.New("boom")
	if AnnotateConnectToNATSError(base) == nil {
		t.Fatalf("expected connect annotation")
	}
	if AnnotatePublishToNATSError("subject", base) == nil {
		t.Fatalf("expected publish annotation")
	}
	if AnnotateFlushNATSPublisherError(base) == nil {
		t.Fatalf("expected flush annotation")
	}
	if AnnotateInitJetStreamContextError(base) == nil {
		t.Fatalf("expected jetstream annotation")
	}
	if AnnotateEnsureConsumerError("stream", "durable", base, base) == nil {
		t.Fatalf("expected ensure consumer annotation")
	}
	if AnnotateCreatePullSubscriberError("subject", "durable", base) == nil {
		t.Fatalf("expected create subscriber annotation")
	}
}
