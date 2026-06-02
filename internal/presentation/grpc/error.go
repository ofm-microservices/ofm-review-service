package grpc

import "errors"

var (
	ErrNilReviewService = errors.New("review service is nil")
	ErrNilReviewMapper  = errors.New("review mapper is nil")
	ErrNilLogger        = errors.New("logger is nil")
)
