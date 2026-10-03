package db

import (
	"errors"
	"testing"
)

func TestAnnotatePostgreSQLErrorHelpers(t *testing.T) {
	base := errors.New("boom")
	if AnnotateResolveMigrationsPathError(base) == nil {
		t.Fatalf("expected resolve annotation")
	}
	if AnnotateCreateMigratorError(base) == nil {
		t.Fatalf("expected create migrator annotation")
	}
	if AnnotateRunMigrationsError(base) == nil {
		t.Fatalf("expected run migrations annotation")
	}
	if AnnotateOpenDBError(base) == nil {
		t.Fatalf("expected open db annotation")
	}
	if AnnotatePingDBError(base) == nil {
		t.Fatalf("expected ping db annotation")
	}
}
