package repository

import (
	"testing"
	"time"

	"review-service/internal/infra/read/redis/model"
)

func TestMapZSetMemberToDomain(t *testing.T) {
	raw := `{"review_id":"review-1","gig_id":"gig-1","content":"c","buyer_id":"buyer-1","rating":5,"author":{"user_id":"buyer-1","username":"user-1","display_name":"User One","avatar_url":"https://example.test/avatar.png","avatar_id":"avatar-1"},"created_at":"2026-05-30T20:10:34.421909Z"}`
	review, err := MapZSetMemberToDomain(raw, 1234)
	if err != nil {
		t.Fatalf("MapZSetMemberToDomain returned error: %v", err)
	}
	if review.Author == nil || review.Author.AvatarID != "avatar-1" {
		t.Fatalf("expected author to be restored, got %#v", review.Author)
	}
	if review.CreatedAt.IsZero() {
		t.Fatalf("expected created at to be parsed")
	}
}

func TestMapZSetMemberToDomainUsesScoreFallback(t *testing.T) {
	raw := `{"review_id":"review-1","gig_id":"gig-1","content":"c","buyer_id":"buyer-1","rating":5}`
	review, err := MapZSetMemberToDomain(raw, 123456)
	if err != nil {
		t.Fatalf("MapZSetMemberToDomain returned error: %v", err)
	}
	if got := review.CreatedAt; !got.Equal(time.UnixMicro(123456).UTC()) {
		t.Fatalf("expected score fallback timestamp, got %s", got)
	}
}

func TestMapZSetMemberToDomainRejectsInvalidMember(t *testing.T) {
	if _, err := MapZSetMemberToDomain(123, 1); err != ErrInvalidReviewMember {
		t.Fatalf("expected invalid member error, got %v", err)
	}
}

func TestParseReviewCreatedAt(t *testing.T) {
	got := ParseReviewCreatedAt("2026-05-30T20:10:34.421909Z", 99)
	if got.IsZero() || got.Location() != time.UTC {
		t.Fatalf("expected parsed UTC time, got %s", got)
	}
	fallback := ParseReviewCreatedAt("broken", 99)
	if !fallback.Equal(time.UnixMicro(99).UTC()) {
		t.Fatalf("expected fallback from score, got %s", fallback)
	}
}

func TestEncodeDecodeCursor(t *testing.T) {
	encoded := EncodeCursor(123456)
	score, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor returned error: %v", err)
	}
	if score != 123456 {
		t.Fatalf("expected score 123456, got %v", score)
	}
	if _, err := DecodeCursor("broken"); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestReviewCacheDefaultAuthorIsNil(t *testing.T) {
	cache := model.ReviewCache{ID: "review-1", GigID: "gig-1"}
	if cache.Author != nil {
		t.Fatalf("expected nil author by default")
	}
}
