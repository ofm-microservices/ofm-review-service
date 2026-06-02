package repository

import (
	"database/sql"
	"errors"
	domain "review-service/internal/domain"
)

// PgErrorTranslator converts sql/Yugabyte errors into domain-aware repository
// errors.
type PgErrorTranslator struct{}

// NewPgErrorTranslator constructs the default Yugabyte error translator.
func NewPgErrorTranslator() DBErrorTranslator {
	return &PgErrorTranslator{}
}

func (t *PgErrorTranslator) TranslateCreateReviewError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return AnnotateDomainError(domain.ErrReviewNotFound, err)
	}
	return AnnotateCreateReviewError(err)
}

func (t *PgErrorTranslator) TranslateFindReviewError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return AnnotateDomainError(domain.ErrReviewNotFound, err)
	}
	return AnnotateFindReviewError(err)
}

func (t *PgErrorTranslator) TranslateDeleteReviewError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return AnnotateDomainError(domain.ErrReviewNotFound, err)
	}
	return AnnotateDeleteReviewError(err)
}
