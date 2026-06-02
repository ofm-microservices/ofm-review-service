package review

import "errors"

var (
	ErrInvalidReviewID = errors.New("invalid review id")
	ErrInvalidOrderID  = errors.New("invalid order id")
	ErrInvalidGigID    = errors.New("invalid gig id")
	// ErrInvalidSellerID indicates a malformed seller-side review owner identifier.
	ErrInvalidSellerID = errors.New("invalid seller id")
	// ErrInvalidUsername indicates a malformed public username handle.
	ErrInvalidUsername      = errors.New("invalid username")
	ErrInvalidContent       = errors.New("invalid review content")
	ErrInvalidBuyerID       = errors.New("invalid buyer id")
	ErrReviewNotFound       = errors.New("review not found")
	ErrReviewNotCompleted   = errors.New("order not completed")
	ErrReviewOwnerMismatch  = errors.New("review owner mismatch")
	ErrReviewAlreadyExists  = errors.New("review already exists")
	ErrFailedToCreateReview = errors.New("failed to create review")
	ErrFailedToFindReview   = errors.New("failed to find review")
	ErrFailedToDeleteReview = errors.New("failed to delete review")
)
