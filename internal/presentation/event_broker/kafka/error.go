package kafka

import "errors"

var (
	ErrNilEventBroker             = errors.New("event broker is nil")
	ErrNilReviewService           = errors.New("review service is nil")
	ErrNilReviewReadRepository    = errors.New("review read repository is nil")
	ErrNilReviewWindowCoordinator = errors.New("review window coordinator is nil")
	ErrNilUserPreviewClient       = errors.New("user preview client is nil")
	ErrNilConfig                  = errors.New("config is nil")
	ErrNilLogger                  = errors.New("logger is nil")
)
