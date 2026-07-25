package repository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	domain "review-service/internal/domain"
	"review-service/internal/infra/write/yugabyte/mapper"
	"review-service/internal/infra/write/yugabyte/model"

	"github.com/jmoiron/sqlx"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metrics"
)

type repo struct {
	db         *sqlx.DB
	translator DBErrorTranslator
	log        logging.Logger
}

// New constructs the Yugabyte-backed review repository.
func New(db *sqlx.DB, translator DBErrorTranslator, log logging.Logger) (domain.ReviewRepository, error) {
	if db == nil {
		return nil, ErrNilYugaByteDB
	}
	if translator == nil {
		return nil, ErrNilDBErrorTranslator
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	return &repo{db: db, translator: translator, log: log.With(logging.String("module", "yugabyte-repository"))}, nil
}

func (r *repo) Create(ctx context.Context, params domain.CreateReviewParams) (*domain.Review, error) {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveDB("yugabyte", "create", "reviews", status, time.Since(started)) }()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		status = "error"
		return nil, r.translator.TranslateCreateReviewError(err)
	}
	defer func() { _ = tx.Rollback() }()

	var existing model.ReviewRow
	if err := tx.QueryRowContext(ctx, selectReviewByOrderIDQuery, params.OrderID).Scan(&existing.ID, &existing.OrderID, &existing.GigID, &existing.Content, &existing.BuyerID, &existing.BuyerUsername, &existing.SellerID, &existing.SellerUsername, &existing.Rating, &existing.CreatedAt, &existing.UpdatedAt); err == nil {
		_ = tx.Commit()
		return mapper.MapReviewRowToDomain(existing), nil
	}

	now := time.Now().UTC()
	var row model.ReviewRow
	if err := tx.QueryRowContext(ctx, insertReviewQuery, params.ID, params.OrderID, params.GigID, params.Content, params.BuyerID, params.BuyerUsername, params.SellerID, params.SellerUsername, params.Rating).Scan(&row.ID, &row.OrderID, &row.GigID, &row.Content, &row.BuyerID, &row.BuyerUsername, &row.SellerID, &row.SellerUsername, &row.Rating, &row.CreatedAt, &row.UpdatedAt); err != nil {
		status = "error"
		return nil, r.translator.TranslateCreateReviewError(err)
	}
	if row.CreatedAt.IsZero() {
		row.CreatedAt = now
		row.UpdatedAt = now
	}
	if err := tx.Commit(); err != nil {
		status = "error"
		return nil, r.translator.TranslateCreateReviewError(err)
	}
	review := mapper.MapReviewRowToDomain(row)
	review.OrderID = params.OrderID
	return review, nil
}

func (r *repo) GetByID(ctx context.Context, reviewID string) (*domain.Review, error) {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveDB("yugabyte", "get_by_id", "reviews", status, time.Since(started)) }()

	var row model.ReviewRow
	if err := r.db.QueryRowContext(ctx, selectReviewByIDQuery, reviewID).Scan(&row.ID, &row.OrderID, &row.GigID, &row.Content, &row.BuyerID, &row.BuyerUsername, &row.SellerID, &row.SellerUsername, &row.Rating, &row.CreatedAt, &row.UpdatedAt); err != nil {
		status = "error"
		return nil, r.translator.TranslateFindReviewError(err)
	}
	return mapper.MapReviewRowToDomain(row), nil
}

func (r *repo) GetByOrderID(ctx context.Context, orderID string) (*domain.Review, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "get_by_order_id", "reviews", status, time.Since(started))
	}()

	var row model.ReviewRow
	if err := r.db.QueryRowContext(ctx, selectReviewByOrderIDQuery, orderID).Scan(&row.ID, &row.OrderID, &row.GigID, &row.Content, &row.BuyerID, &row.BuyerUsername, &row.SellerID, &row.SellerUsername, &row.Rating, &row.CreatedAt, &row.UpdatedAt); err != nil {
		status = "error"
		return nil, r.translator.TranslateFindReviewError(err)
	}
	return mapper.MapReviewRowToDomain(row), nil
}

func (r *repo) GetSellerIDByUsername(ctx context.Context, username string) (string, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "get_seller_id_by_username", "reviews", status, time.Since(started))
	}()

	var sellerID sql.NullString
	if err := r.db.QueryRowContext(ctx, selectSellerIDByUsernameQuery, strings.TrimSpace(username)).Scan(&sellerID); err != nil {
		status = "error"
		return "", r.translator.TranslateFindReviewError(err)
	}
	if !sellerID.Valid || strings.TrimSpace(sellerID.String) == "" {
		status = "error"
		return "", AnnotateDomainError(domain.ErrReviewNotFound, sql.ErrNoRows)
	}
	return strings.TrimSpace(sellerID.String), nil
}

func (r *repo) ListByGigID(ctx context.Context, query domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "list_by_gig_id", "reviews", status, time.Since(started))
	}()
	return r.ListByIndex(ctx, selectReviewsByGigIDQuery, selectReviewsByGigIDCursorQuery, query.GigID, query.Cursor, query.Limit)
}

func (r *repo) ListBySellerID(ctx context.Context, query domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "list_by_seller_id", "reviews", status, time.Since(started))
	}()
	return r.ListByIndex(ctx, selectReviewsBySellerIDQuery, selectReviewsBySellerIDCursorQuery, query.SellerID, query.Cursor, query.Limit)
}

func (r *repo) GetGigRatingSummary(ctx context.Context, gigID string) (*domain.RatingSummary, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "get_gig_rating_summary", "reviews", status, time.Since(started))
	}()

	var avg float64
	var total, stars5, stars4, stars3, stars2, stars1 int64
	if err := r.db.QueryRowContext(ctx, selectGigRatingSummaryQuery, gigID).Scan(&avg, &total, &stars5, &stars4, &stars3, &stars2, &stars1); err != nil {
		status = "error"
		return nil, r.translator.TranslateFindReviewError(err)
	}
	if total == 0 {
		status = "error"
		return nil, AnnotateDomainError(domain.ErrReviewNotFound, sql.ErrNoRows)
	}
	return buildRatingSummary(avg, total, stars5, stars4, stars3, stars2, stars1), nil
}

func (r *repo) GetSellerRatingSummary(ctx context.Context, sellerID string) (*domain.RatingSummary, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "get_seller_rating_summary", "reviews", status, time.Since(started))
	}()

	var avg float64
	var total, stars5, stars4, stars3, stars2, stars1 int64
	if err := r.db.QueryRowContext(ctx, selectSellerRatingSummaryQuery, sellerID).Scan(&avg, &total, &stars5, &stars4, &stars3, &stars2, &stars1); err != nil {
		status = "error"
		return nil, r.translator.TranslateFindReviewError(err)
	}
	if total == 0 {
		status = "error"
		return nil, AnnotateDomainError(domain.ErrReviewNotFound, sql.ErrNoRows)
	}
	return buildRatingSummary(avg, total, stars5, stars4, stars3, stars2, stars1), nil
}

func (r *repo) ListGigRatingSummaries(ctx context.Context) ([]domain.GigRatingSummary, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "list_gig_rating_summaries", "reviews", status, time.Since(started))
	}()

	var rows []model.RatingSummaryRow
	if err := r.db.SelectContext(ctx, &rows, selectGigRatingSummariesQuery); err != nil {
		status = "error"
		return nil, r.translator.TranslateFindReviewError(err)
	}
	return mapper.MapGigRatingSummaryRowsToDomain(rows), nil
}

func (r *repo) ListSellerRatingSummaries(ctx context.Context) ([]domain.SellerRatingSummary, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "list_seller_rating_summaries", "reviews", status, time.Since(started))
	}()

	var rows []model.RatingSummaryRow
	if err := r.db.SelectContext(ctx, &rows, selectSellerRatingSummariesQuery); err != nil {
		status = "error"
		return nil, r.translator.TranslateFindReviewError(err)
	}
	return mapper.MapSellerRatingSummaryRowsToDomain(rows), nil
}

func (r *repo) DeleteByID(ctx context.Context, reviewID string) error {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveDB("yugabyte", "delete_by_id", "reviews", status, time.Since(started)) }()

	if _, err := r.db.ExecContext(ctx, deleteReviewByIDQuery, reviewID); err != nil {
		status = "error"
		return r.translator.TranslateDeleteReviewError(err)
	}
	return nil
}

func (r *repo) ListByIndex(ctx context.Context, firstQuery, cursorQuery, ownerID, cursor string, limit int) (*domain.ListReviewsResult, error) {
	if limit <= 0 {
		limit = 1
	}

	cursor = strings.TrimSpace(cursor)

	if cursor == "" {
		var rows []model.ReviewRow
		if err := r.db.SelectContext(ctx, &rows, firstQuery, ownerID, limit+1); err != nil {
			return nil, r.translator.TranslateFindReviewError(err)
		}
		return r.RowsToResult(rows, limit)
	}

	createdAt, reviewID, err := DecodeReviewCursor(cursor)
	if err != nil {
		return nil, err
	}
	var rows []model.ReviewRow
	if err := r.db.SelectContext(ctx, &rows, cursorQuery, ownerID, createdAt, reviewID, limit+1); err != nil {
		return nil, r.translator.TranslateFindReviewError(err)
	}
	return r.RowsToResult(rows, limit)
}

func (r *repo) RowsToResult(rows []model.ReviewRow, limit int) (*domain.ListReviewsResult, error) {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	reviews := make([]*domain.Review, 0, len(rows))
	cursor := ""
	for _, row := range rows {
		review := mapper.MapReviewRowToDomain(row)
		reviews = append(reviews, review)
		cursor = EncodeReviewCursor(review.CreatedAt.UTC().UnixMicro(), review.ID)
	}
	if !hasMore {
		cursor = ""
	}
	return &domain.ListReviewsResult{
		Reviews: reviews,
		Cursor:  cursor,
		HasMore: hasMore,
	}, nil
}

func buildRatingSummary(avg float64, total, stars5, stars4, stars3, stars2, stars1 int64) *domain.RatingSummary {
	return &domain.RatingSummary{
		RatingAvg:    avg,
		TotalReviews: total,
		Stars5:       stars5,
		Stars4:       stars4,
		Stars3:       stars3,
		Stars2:       stars2,
		Stars1:       stars1,
	}
}

func EncodeReviewCursor(createdAtMicros int64, reviewID string) string {
	raw := strconv.FormatInt(createdAtMicros, 10) + ":" + reviewID
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeReviewCursor(cursor string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", ErrInvalidCursor
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return time.Time{}, "", ErrInvalidCursor
	}
	micros, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, "", ErrInvalidCursor
	}
	return time.UnixMicro(micros).UTC(), parts[1], nil
}
