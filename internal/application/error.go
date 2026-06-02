package service

import "errors"

var (
	ErrNilReviewRepository     = errors.New("review repository is nil")
	ErrNilReviewReadRepository = errors.New("review read repository is nil")
	ErrNilOrderLookupClient    = errors.New("order lookup client is nil")
	ErrNilUserPreviewClient    = errors.New("user preview client is nil")
	ErrNilFileURLClient        = errors.New("file url client is nil")
	ErrNilReviewPublisher      = errors.New("review publisher is nil")
	ErrNilReviewService        = errors.New("review service is nil")
	ErrNilLogger               = errors.New("logger is nil")
	ErrNilCursorCodec          = errors.New("cursor codec is nil")
	ErrInvalidPaginationConfig = errors.New("invalid pagination config")
	ErrQueueDurabilityDisabled = errors.New("review queue durability is disabled")
	ErrQueueLeaseLost          = errors.New("review queue lease was lost")
)
