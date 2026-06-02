package repository

import (
	"errors"
	"fmt"

	domain "review-service/internal/domain"
)

var (
	ErrNilYugaByteDB        = errors.New("yugabyte db is nil")
	ErrNilDBErrorTranslator = errors.New("db error translator is nil")
	ErrNilLogger            = errors.New("logger is nil")
	ErrInvalidCursor        = errors.New("invalid review cursor")
)

const domainWrapFormat = "%w: %v"

// AnnotateCreateReviewError annotates review insert failures.
func AnnotateCreateReviewError(err error) error {
	return fmt.Errorf(domainWrapFormat, domain.ErrFailedToCreateReview, err)
}

// AnnotateFindReviewError annotates review lookup failures.
func AnnotateFindReviewError(err error) error {
	return fmt.Errorf(domainWrapFormat, domain.ErrFailedToFindReview, err)
}

// AnnotateDeleteReviewError annotates review delete failures.
func AnnotateDeleteReviewError(err error) error {
	return fmt.Errorf(domainWrapFormat, domain.ErrFailedToDeleteReview, err)
}

// AnnotateDomainError preserves the domain error while attaching the original
// database cause for logging and debugging.
func AnnotateDomainError(domainErr, err error) error {
	return fmt.Errorf("%w: %v", domainErr, err)
}
