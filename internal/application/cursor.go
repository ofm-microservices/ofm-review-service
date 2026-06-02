package service

import (
	"strings"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"review-service/internal/application/cursordb"
	domain "review-service/internal/domain"
)

// ReviewCursorState carries the opaque pagination state for review list reads.
type ReviewCursorState struct {
	Scope     string `json:"scope"`
	Window    int    `json:"window"`
	Page      int    `json:"page"`
	LastScore int64  `json:"last_score"`
	LastID    string `json:"last_id"`
}

// DecodeReviewCursor decodes a cursor token into a review pagination state.
func (s *reviewService) DecodeReviewCursor(raw, scope string) (ReviewCursorState, error) {
	if strings.TrimSpace(raw) == "" {
		return ReviewCursorState{Scope: scope, Window: 0, Page: 1}, nil
	}
	var state ReviewCursorState
	if err := s.cursor.Decode(raw, &state); err != nil {
		return ReviewCursorState{}, err
	}
	if state.Scope != scope {
		return ReviewCursorState{}, domain.ErrInvalidReviewID
	}
	if state.Page <= 0 {
		state.Page = 1
	}
	return state, nil
}

// EncodeReviewCursor encodes a review cursor state into an opaque token.
func (s *reviewService) EncodeReviewCursor(state ReviewCursorState) string {
	token, err := s.cursor.Encode(state)
	if err != nil {
		s.log.Error("encode review cursor failed",
			logging.Operation("review.cursor.encode"),
			logging.String("scope", state.Scope),
			logging.Int("window", state.Window),
			logging.Int("page", state.Page),
			logging.Err(err),
		)
		return ""
	}
	return token
}

// Next advances the cursor state after a page has been served.
func (s ReviewCursorState) Next(lastScore int64, lastID string, pagesPerWindow int) ReviewCursorState {
	s.LastScore = lastScore
	s.LastID = lastID
	if pagesPerWindow <= 0 {
		pagesPerWindow = 1
	}
	pageInWindow := s.Page % pagesPerWindow
	if pageInWindow == 0 {
		s.Window++
		s.Page = 1
		return s
	}
	s.Page++
	return s
}

// DBCursor returns the canonical cursor used for the YugabyteDB read path.
func (s ReviewCursorState) DBCursor() string {
	if s.LastScore == 0 && strings.TrimSpace(s.LastID) == "" {
		return ""
	}
	return cursordb.New().Encode(s.LastScore, s.LastID)
}
