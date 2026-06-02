package nats

import "fmt"

// AnnotateEnsureStreamError annotates add-or-update stream bootstrap failures.
func AnnotateEnsureStreamError(streamName string, addErr, updateErr error) error {
	return fmt.Errorf("ensure stream %q: add stream failed: %w; update stream failed: %v", streamName, addErr, updateErr)
}

// AnnotateConnectToNATSError annotates NATS connection failures.
func AnnotateConnectToNATSError(err error) error {
	return fmt.Errorf("connect to nats: %w", err)
}

// AnnotateInitJetStreamContextError annotates JetStream context initialization failures.
func AnnotateInitJetStreamContextError(err error) error {
	return fmt.Errorf("init jetstream context: %w", err)
}
