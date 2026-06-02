package repository

import (
	"testing"
	"time"

	"review-service/internal/infra/write/yugabyte/model"
)

func TestEncodeDecodeReviewCursor(t *testing.T) {
	encoded := EncodeReviewCursor(123456, "review-1")
	gotTime, gotID, err := DecodeReviewCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeReviewCursor returned error: %v", err)
	}
	if !gotTime.Equal(time.UnixMicro(123456).UTC()) || gotID != "review-1" {
		t.Fatalf("unexpected cursor decode: %s %s", gotTime, gotID)
	}
	if _, _, err := DecodeReviewCursor("broken"); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestRowsToResultPaginatesAndBuildsCursor(t *testing.T) {
	repo := &repo{}
	rows := []model.ReviewRow{
		{ID: "r1", OrderID: "o1", GigID: "g1", Content: "c1", BuyerID: "b1", Rating: 5, CreatedAt: time.UnixMicro(1000).UTC(), UpdatedAt: time.UnixMicro(1000).UTC()},
		{ID: "r2", OrderID: "o2", GigID: "g1", Content: "c2", BuyerID: "b2", Rating: 4, CreatedAt: time.UnixMicro(2000).UTC(), UpdatedAt: time.UnixMicro(2000).UTC()},
		{ID: "r3", OrderID: "o3", GigID: "g1", Content: "c3", BuyerID: "b3", Rating: 3, CreatedAt: time.UnixMicro(3000).UTC(), UpdatedAt: time.UnixMicro(3000).UTC()},
	}

	res, err := repo.RowsToResult(rows, 2)
	if err != nil {
		t.Fatalf("RowsToResult returned error: %v", err)
	}
	if len(res.Reviews) != 2 {
		t.Fatalf("expected 2 reviews, got %d", len(res.Reviews))
	}
	if !res.HasMore {
		t.Fatalf("expected hasMore")
	}
	if res.Cursor == "" {
		t.Fatalf("expected cursor")
	}
}

func TestRowsToResultFinalPageClearsCursor(t *testing.T) {
	repo := &repo{}
	rows := []model.ReviewRow{
		{ID: "r1", OrderID: "o1", GigID: "g1", Content: "c1", BuyerID: "b1", Rating: 5, CreatedAt: time.UnixMicro(1000).UTC(), UpdatedAt: time.UnixMicro(1000).UTC()},
	}

	res, err := repo.RowsToResult(rows, 2)
	if err != nil {
		t.Fatalf("RowsToResult returned error: %v", err)
	}
	if res.HasMore {
		t.Fatalf("expected no more")
	}
	if res.Cursor != "" {
		t.Fatalf("expected empty cursor, got %q", res.Cursor)
	}
}
