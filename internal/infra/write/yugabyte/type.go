package repository

// DBErrorTranslator maps storage-driver failures into domain-aware repository
// errors.
type DBErrorTranslator interface {
	TranslateCreateReviewError(err error) error
	TranslateFindReviewError(err error) error
	TranslateDeleteReviewError(err error) error
}
