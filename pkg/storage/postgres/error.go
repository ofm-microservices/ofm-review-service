package db

import "fmt"

// AnnotateResolveMigrationsPathError annotates relative migration path resolution
// failures.
func AnnotateResolveMigrationsPathError(err error) error {
	return fmt.Errorf("resolve migrations path: %w", err)
}

// AnnotateCreateMigratorError annotates migrate client construction failures.
func AnnotateCreateMigratorError(err error) error {
	return fmt.Errorf("create migrator: %w", err)
}

// AnnotateRunMigrationsError annotates migration execution failures.
func AnnotateRunMigrationsError(err error) error {
	return fmt.Errorf("run migrations: %w", err)
}

// AnnotateOpenDBError annotates SQL driver connection failures.
func AnnotateOpenDBError(err error) error {
	return fmt.Errorf("open db: %w", err)
}

// AnnotatePingDBError annotates health-check ping failures.
func AnnotatePingDBError(err error) error {
	return fmt.Errorf("ping db: %w", err)
}
