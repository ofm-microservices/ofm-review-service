package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	domain "review-service/internal/domain"
	"review-service/internal/infra/read/redis/mapper"
	"review-service/internal/infra/read/redis/model"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metrics"
	"github.com/redis/go-redis/v9"
)

type repo struct {
	rdb *redis.Client
	log logging.Logger
}

// New constructs the Redis-backed review read-model repository.
func New(rdb *redis.Client, log logging.Logger) (domain.ReviewReadRepository, error) {
	if rdb == nil {
		return nil, ErrNilRedisClient
	}
	if log == nil {
		return nil, ErrNilLogger
	}

	return &repo{rdb: rdb, log: log.With(logging.String("module", "redis-repository"))}, nil
}

func (r *repo) UpsertGigRating(ctx context.Context, gigID string, rating int32) error {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("set", "review", status, time.Since(started)) }()

	if strings.TrimSpace(gigID) == "" {
		return ErrInvalidCursor
	}
	if err := r.UpsertRatingAggregate(ctx, GigRatingKey(gigID), rating); err != nil {
		status = "error"
		return err
	}
	return nil
}

func (r *repo) UpsertSellerRating(ctx context.Context, sellerID string, rating int32) error {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("set", "review", status, time.Since(started)) }()

	if strings.TrimSpace(sellerID) == "" {
		return ErrInvalidCursor
	}
	if err := r.UpsertRatingAggregate(ctx, SellerRatingKey(sellerID), rating); err != nil {
		status = "error"
		return err
	}
	return nil
}

func (r *repo) UpsertSellerRatingByUsername(ctx context.Context, username string, rating int32) error {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("set", "review", status, time.Since(started)) }()

	if strings.TrimSpace(username) == "" {
		return ErrInvalidCursor
	}
	if err := r.UpsertRatingAggregate(ctx, SellerRatingByUsernameKey(username), rating); err != nil {
		status = "error"
		return err
	}
	return nil
}

func (r *repo) GetGigRatingSummary(ctx context.Context, gigID string) (*domain.RatingSummary, error) {
	return r.getRatingSummary(ctx, GigRatingKey(gigID))
}

func (r *repo) GetSellerRatingSummary(ctx context.Context, sellerID string) (*domain.RatingSummary, error) {
	return r.getRatingSummary(ctx, SellerRatingKey(sellerID))
}

func (r *repo) GetSellerRatingSummaryByUsername(ctx context.Context, username string) (*domain.RatingSummary, error) {
	return r.getRatingSummary(ctx, SellerRatingByUsernameKey(username))
}

func (r *repo) GetSellerIDByUsername(ctx context.Context, username string) (string, error) {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("get", "review", status, time.Since(started)) }()

	key := SellerIDByUsernameKey(username)
	raw, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		status = "error"
		if err == redis.Nil {
			return "", domain.ErrReviewNotFound
		}
		return "", AnnotateSetReviewCacheError(key, err)
	}
	return strings.TrimSpace(raw), nil
}

func (r *repo) SetGigRating(ctx context.Context, gigID string, summary domain.RatingSummary) error {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("set", "review", status, time.Since(started)) }()

	if strings.TrimSpace(gigID) == "" {
		return ErrInvalidCursor
	}
	return r.setRatingAggregate(ctx, GigRatingKey(gigID), summary)
}

func (r *repo) SetSellerRating(ctx context.Context, sellerID string, summary domain.RatingSummary) error {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("set", "review", status, time.Since(started)) }()

	if strings.TrimSpace(sellerID) == "" {
		return ErrInvalidCursor
	}
	return r.setRatingAggregate(ctx, SellerRatingKey(sellerID), summary)
}

func (r *repo) SetSellerRatingByUsername(ctx context.Context, username string, summary domain.RatingSummary) error {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("set", "review", status, time.Since(started)) }()

	if strings.TrimSpace(username) == "" {
		return ErrInvalidCursor
	}
	return r.setRatingAggregate(ctx, SellerRatingByUsernameKey(username), summary)
}

func (r *repo) SetSellerIDByUsername(ctx context.Context, username, sellerID string) error {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("set", "review", status, time.Since(started)) }()

	username = strings.TrimSpace(username)
	sellerID = strings.TrimSpace(sellerID)
	if username == "" || sellerID == "" {
		return ErrInvalidCursor
	}
	if err := r.rdb.Set(ctx, SellerIDByUsernameKey(username), sellerID, 0).Err(); err != nil {
		status = "error"
		return AnnotateSetReviewCacheError(SellerIDByUsernameKey(username), err)
	}
	return nil
}

func (r *repo) UpsertGigWindow(ctx context.Context, gigID string, window int, reviews []*domain.Review, hasMore bool, ttl time.Duration) error {
	return r.UpsertWindow(ctx, GigReviewsWindowKey(gigID, window), reviews, ttl)
}

func (r *repo) UpsertSellerWindow(ctx context.Context, sellerID string, window int, reviews []*domain.Review, hasMore bool, ttl time.Duration) error {
	return r.UpsertWindow(ctx, SellerReviewsWindowKey(sellerID, window), reviews, ttl)
}

func (r *repo) ListByGigID(ctx context.Context, query domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return r.ListByKey(ctx, GigReviewsKey(query.GigID), query.Cursor, query.Limit)
}

func (r *repo) ListBySellerID(ctx context.Context, query domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return r.ListByKey(ctx, SellerReviewsKey(query.SellerID), query.Cursor, query.Limit)
}

func (r *repo) ListByGigWindow(ctx context.Context, gigID string, window int) (*domain.ListReviewsResult, error) {
	return r.ListWindow(ctx, GigReviewsWindowKey(gigID, window))
}

func (r *repo) ListBySellerWindow(ctx context.Context, sellerID string, window int) (*domain.ListReviewsResult, error) {
	return r.ListWindow(ctx, SellerReviewsWindowKey(sellerID, window))
}

func (r *repo) DeleteByID(ctx context.Context, reviewID string) error {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("del", "review", status, time.Since(started)) }()

	key := ReviewCacheKey(reviewID)
	if err := r.rdb.Del(ctx, key).Err(); err != nil {
		status = "error"
		r.log.Error("delete review cache failed",
			logging.Operation("redis.review.delete_by_id"),
			logging.Attempt(1),
			logging.Retryable(false),
			logging.DurationMS(time.Since(started)),
			logging.String("key", key),
			logging.Err(err),
		)
		return AnnotateDeleteReviewCacheError(key, err)
	}

	return nil
}

// ReviewCacheKey builds the Redis key used for the review read model.
func ReviewCacheKey(reviewID string) string {
	return fmt.Sprintf("review:%s", reviewID)
}

// ReviewIndexMember builds the stable ZSET member used for cursor pagination.
func ReviewIndexMember(createdAt time.Time, reviewID string) string {
	return fmt.Sprintf("%020d:%s", createdAt.UTC().UnixNano(), reviewID)
}

// GigReviewsKey builds the ZSET key used for gig review pagination.
func GigReviewsKey(gigID string) string {
	return fmt.Sprintf("gig:reviews:%s", gigID)
}

// SellerReviewsKey builds the ZSET key used for seller review pagination.
func SellerReviewsKey(sellerID string) string {
	return fmt.Sprintf("seller:reviews:%s", sellerID)
}

// GigReviewsWindowKey builds the Redis key used for a gig review cache window.
func GigReviewsWindowKey(gigID string, window int) string {
	return fmt.Sprintf("gig:reviews:%s:window:%d", gigID, window)
}

// SellerReviewsWindowKey builds the Redis key used for a seller review cache window.
func SellerReviewsWindowKey(sellerID string, window int) string {
	return fmt.Sprintf("seller:reviews:%s:window:%d", sellerID, window)
}

// GigRatingKey builds the aggregate key used for gig rating summaries.
func GigRatingKey(gigID string) string {
	return fmt.Sprintf("gig:rating:%s", gigID)
}

// SellerRatingKey builds the aggregate key used for seller rating summaries.
func SellerRatingKey(sellerID string) string {
	return fmt.Sprintf("seller:rating:%s", sellerID)
}

// SellerRatingByUsernameKey builds the aggregate key used for seller rating summaries resolved by username.
func SellerRatingByUsernameKey(username string) string {
	return fmt.Sprintf("seller:rating:username:%s", username)
}

// SellerIDByUsernameKey builds the cache key used for seller ID lookup by username.
func SellerIDByUsernameKey(username string) string {
	return fmt.Sprintf("user:%s", username)
}

func (r *repo) ReviewIndexMemberExists(ctx context.Context, key, reviewID string) (bool, error) {
	_ = ctx
	_ = key
	_ = reviewID
	return false, nil
}

func (r *repo) ListByKey(ctx context.Context, key, cursor string, limit int) (*domain.ListReviewsResult, error) {
	if limit <= 0 {
		limit = 1
	}

	max := "+inf"
	if cursor = strings.TrimSpace(cursor); cursor != "" {
		score, err := DecodeCursor(cursor)
		if err != nil {
			return nil, err
		}
		max = fmt.Sprintf("(%s", strconv.FormatFloat(score, 'f', -1, 64))
	}

	rows, err := r.rdb.ZRevRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Max:    max,
		Min:    "-inf",
		Offset: 0,
		Count:  int64(limit + 1),
	}).Result()
	if err != nil {
		return nil, AnnotateSetReviewCacheError(key, err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	reviews := make([]*domain.Review, 0, len(rows))
	nextCursor := ""
	for _, row := range rows {
		review, err := MapZSetMemberToDomain(row.Member, row.Score)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
		nextCursor = EncodeCursor(row.Score)
	}
	if !hasMore {
		nextCursor = ""
	}

	return &domain.ListReviewsResult{
		Reviews: reviews,
		Cursor:  nextCursor,
		HasMore: hasMore,
	}, nil
}

func EncodeCursor(score float64) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatFloat(score, 'f', 0, 64)))
}

func DecodeCursor(cursor string) (float64, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	score, err := strconv.ParseFloat(string(raw), 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(score) || math.IsInf(score, 0) {
		return 0, ErrInvalidCursor
	}
	return score, nil
}

func MapZSetMemberToDomain(member any, score float64) (*domain.Review, error) {
	raw, ok := member.(string)
	if !ok {
		return nil, ErrInvalidReviewMember
	}
	var cache model.ReviewCache
	if err := json.Unmarshal([]byte(raw), &cache); err != nil {
		return nil, AnnotateUnmarshalReviewCacheError(err)
	}
	review := &domain.Review{
		ID:             cache.ID,
		GigID:          cache.GigID,
		Content:        cache.Content,
		BuyerID:        cache.BuyerID,
		SellerUsername: cache.SellerUsername,
		Rating:         cache.Rating,
		CreatedAt:      ParseReviewCreatedAt(cache.CreatedAt, score),
	}
	if cache.Author != nil {
		review.Author = &domain.ReviewAuthor{
			UserID:      cache.Author.UserID,
			Username:    cache.Author.Username,
			DisplayName: cache.Author.DisplayName,
			AvatarURL:   cache.Author.AvatarURL,
			AvatarID:    cache.Author.AvatarID,
		}
	}
	return review, nil
}

func ParseReviewCreatedAt(raw string, score float64) time.Time {
	if strings.TrimSpace(raw) != "" {
		if ts, err := time.Parse(time.RFC3339Nano, raw); err == nil {
			return ts.UTC()
		}
	}
	return time.UnixMicro(int64(score)).UTC()
}

func (r *repo) UpsertRatingAggregate(ctx context.Context, key string, rating int32) error {
	var agg model.RatingAggregate
	if raw, err := r.rdb.Get(ctx, key).Bytes(); err == nil {
		_ = json.Unmarshal(raw, &agg)
	}
	agg.TotalReviews++
	switch rating {
	case 5:
		agg.Stars5++
	case 4:
		agg.Stars4++
	case 3:
		agg.Stars3++
	case 2:
		agg.Stars2++
	default:
		agg.Stars1++
	}
	total := float64(agg.TotalReviews)
	sum := float64(agg.Stars5*5 + agg.Stars4*4 + agg.Stars3*3 + agg.Stars2*2 + agg.Stars1)
	if total > 0 {
		agg.RatingAvg = sum / total
	}
	payload, err := json.Marshal(agg)
	if err != nil {
		return AnnotateMarshalReviewCacheError(err)
	}
	if err := r.rdb.Set(ctx, key, payload, 0).Err(); err != nil {
		return AnnotateSetReviewCacheError(key, err)
	}
	return nil
}

func (r *repo) setRatingAggregate(ctx context.Context, key string, summary domain.RatingSummary) error {
	payload, err := json.Marshal(model.RatingAggregate{
		RatingAvg:    summary.RatingAvg,
		TotalReviews: summary.TotalReviews,
		Stars5:       summary.Stars5,
		Stars4:       summary.Stars4,
		Stars3:       summary.Stars3,
		Stars2:       summary.Stars2,
		Stars1:       summary.Stars1,
	})
	if err != nil {
		return AnnotateMarshalReviewCacheError(err)
	}
	if err := r.rdb.Set(ctx, key, payload, 0).Err(); err != nil {
		return AnnotateSetReviewCacheError(key, err)
	}
	return nil
}

func (r *repo) getRatingSummary(ctx context.Context, key string) (*domain.RatingSummary, error) {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("get", "review", status, time.Since(started)) }()

	raw, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		status = "error"
		if err == redis.Nil {
			return nil, domain.ErrReviewNotFound
		}
		return nil, AnnotateSetReviewCacheError(key, err)
	}

	var agg model.RatingAggregate
	if err := json.Unmarshal(raw, &agg); err != nil {
		status = "error"
		return nil, AnnotateUnmarshalReviewCacheError(err)
	}
	if agg.TotalReviews == 0 {
		_ = r.rdb.Del(ctx, key).Err()
		return nil, domain.ErrReviewNotFound
	}
	return &domain.RatingSummary{
		RatingAvg:    agg.RatingAvg,
		TotalReviews: agg.TotalReviews,
		Stars5:       agg.Stars5,
		Stars4:       agg.Stars4,
		Stars3:       agg.Stars3,
		Stars2:       agg.Stars2,
		Stars1:       agg.Stars1,
	}, nil
}

func (r *repo) UpsertWindow(ctx context.Context, key string, reviews []*domain.Review, ttl time.Duration) error {
	if len(reviews) == 0 {
		return nil
	}
	pipe := r.rdb.TxPipeline()
	for _, review := range reviews {
		cache := mapper.MapDomainReviewToCache(review)
		payload, err := json.Marshal(cache)
		if err != nil {
			return AnnotateMarshalReviewCacheError(err)
		}
		score := float64(review.CreatedAt.UTC().UnixMicro())
		pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: string(payload)})
	}
	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return AnnotateSetReviewCacheError(key, err)
	}
	return nil
}

func (r *repo) ListWindow(ctx context.Context, key string) (*domain.ListReviewsResult, error) {
	rows, err := r.rdb.ZRevRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, AnnotateSetReviewCacheError(key, err)
	}
	if len(rows) == 0 {
		return nil, domain.ErrReviewNotFound
	}
	reviews := make([]*domain.Review, 0, len(rows))
	for _, row := range rows {
		review, err := MapZSetMemberToDomain(row.Member, row.Score)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return &domain.ListReviewsResult{Reviews: reviews}, nil
}

// EnqueueWindowJob appends a review job to the owner queue and indexes the
// queue in the active set with one atomic Redis script.
func (r *repo) EnqueueWindowJob(ctx context.Context, queueKey, activeKey string, review *domain.Review) (int64, error) {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("rpush", "review", status, time.Since(started)) }()

	if review == nil {
		return 0, ErrNilReview
	}
	payload, err := json.Marshal(review)
	if err != nil {
		status = "error"
		return 0, AnnotateMarshalReviewCacheError(err)
	}

	depth, err := reviewWindowEnqueueScript.Run(ctx, r.rdb, []string{queueKey, activeKey}, string(payload)).Int64()
	if err != nil {
		status = "error"
		return 0, AnnotateEnqueueReviewWindowJobError(queueKey, err)
	}
	return depth, nil
}

// QueueLength returns the number of queued review jobs for an owner.
func (r *repo) QueueLength(ctx context.Context, queueKey string) (int64, error) {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("llen", "review", status, time.Since(started)) }()

	depth, err := r.rdb.LLen(ctx, queueKey).Result()
	if err != nil {
		status = "error"
		return 0, AnnotateQueueDepthReviewError(queueKey, err)
	}
	return depth, nil
}

// PeekWindowJob returns the queued review at the front of the owner queue
// without removing it.
func (r *repo) PeekWindowJob(ctx context.Context, queueKey string) (*domain.Review, error) {
	raw, err := r.rdb.LIndex(ctx, queueKey, 0).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrReviewNotFound
		}
		return nil, AnnotateQueuePeekReviewError(queueKey, err)
	}
	if strings.TrimSpace(raw) == "" {
		return nil, domain.ErrReviewNotFound
	}

	var review domain.Review
	if err := json.Unmarshal([]byte(raw), &review); err != nil {
		return nil, AnnotateUnmarshalReviewCacheError(err)
	}
	return &review, nil
}

// PopWindowJob removes the queue head after a successful window mutation.
func (r *repo) PopWindowJob(ctx context.Context, queueKey string) (*domain.Review, error) {
	raw, err := r.rdb.LPop(ctx, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrReviewNotFound
		}
		return nil, AnnotateQueuePopReviewError(queueKey, err)
	}
	if strings.TrimSpace(raw) == "" {
		return nil, domain.ErrReviewNotFound
	}

	var review domain.Review
	if err := json.Unmarshal([]byte(raw), &review); err != nil {
		return nil, AnnotateUnmarshalReviewCacheError(err)
	}
	return &review, nil
}

// ActiveQueueKeys returns the set of queue keys that still have queued jobs.
func (r *repo) ActiveQueueKeys(ctx context.Context, activeKey string) ([]string, error) {
	keys, err := r.rdb.SMembers(ctx, activeKey).Result()
	if err != nil {
		return nil, AnnotateActiveQueueReviewError(activeKey, err)
	}
	return keys, nil
}

// RemoveActiveQueueIfEmpty removes a queue key from the active set when the
// queue has no remaining jobs.
func (r *repo) RemoveActiveQueueIfEmpty(ctx context.Context, activeKey, queueKey string) (bool, error) {
	res, err := reviewWindowInactiveCleanupScript.Run(ctx, r.rdb, []string{queueKey, activeKey}).Int64()
	if err != nil {
		return false, AnnotateActiveQueueReviewError(activeKey, err)
	}
	return res == 1, nil
}

// TryAcquireLease acquires the owner lease with SET NX PX.
func (r *repo) TryAcquireLease(ctx context.Context, leaseKey, token string, ttl time.Duration) (bool, error) {
	ok, err := r.rdb.SetNX(ctx, leaseKey, token, ttl).Result()
	if err != nil {
		return false, AnnotateLeaseReviewError(leaseKey, err)
	}
	return ok, nil
}

// RenewLease extends the owner lease only if the caller still owns it.
func (r *repo) RenewLease(ctx context.Context, leaseKey, token string, ttl time.Duration) (bool, error) {
	ttlMs := int64(ttl / time.Millisecond)
	if ttlMs <= 0 {
		ttlMs = 1
	}
	ok, err := reviewWindowLeaseRenewScript.Run(ctx, r.rdb, []string{leaseKey}, token, ttlMs).Int64()
	if err != nil {
		return false, AnnotateLeaseReviewError(leaseKey, err)
	}
	return ok == 1, nil
}

// ReleaseLease deletes the lease only if the caller still owns it.
func (r *repo) ReleaseLease(ctx context.Context, leaseKey, token string) (bool, error) {
	ok, err := reviewWindowLeaseReleaseScript.Run(ctx, r.rdb, []string{leaseKey}, token).Int64()
	if err != nil {
		return false, AnnotateLeaseReviewError(leaseKey, err)
	}
	return ok == 1, nil
}

// LoadWindowSnapshots reads all current review-window boundaries for the owner
// in one Redis round-trip.
func (r *repo) LoadWindowSnapshots(ctx context.Context, ownerPrefix string) ([]domain.WindowSnapshot, error) {
	raw, err := reviewWindowSnapshotScript.Run(ctx, r.rdb, nil, ownerPrefix).Result()
	if err != nil {
		return nil, AnnotateLoadReviewWindowSnapshotsError(ownerPrefix, err)
	}

	parts, ok := raw.([]interface{})
	if !ok {
		return nil, AnnotateLoadReviewWindowSnapshotsError(ownerPrefix, ErrInvalidReviewMember)
	}

	snapshots := make([]domain.WindowSnapshot, 0, len(parts)/6)
	for i := 0; i+5 < len(parts); i += 6 {
		window, err := toInt(parts[i])
		if err != nil {
			return nil, AnnotateLoadReviewWindowSnapshotsError(ownerPrefix, err)
		}
		count, err := toInt(parts[i+1])
		if err != nil {
			return nil, AnnotateLoadReviewWindowSnapshotsError(ownerPrefix, err)
		}
		newestReview, err := toReviewFromCache(parts[i+2], parts[i+3])
		if err != nil {
			return nil, AnnotateLoadReviewWindowSnapshotsError(ownerPrefix, err)
		}
		oldestReview, err := toReviewFromCache(parts[i+4], parts[i+5])
		if err != nil {
			return nil, AnnotateLoadReviewWindowSnapshotsError(ownerPrefix, err)
		}
		snapshots = append(snapshots, domain.WindowSnapshot{
			Window: window,
			Count:  count,
			Newest: domain.ReviewMember{Review: newestReview, Score: toScore(parts[i+3])},
			Oldest: domain.ReviewMember{Review: oldestReview, Score: toScore(parts[i+5])},
		})
	}
	return snapshots, nil
}

// ApplyWindowOpsAndPopJob applies a full review-window carry chain atomically
// and removes the queue head only if it still matches the review being processed.
func (r *repo) ApplyWindowOpsAndPopJob(ctx context.Context, queueKey, leaseKey, token, ownerPrefix, reviewID string, ops []domain.WindowOp, ttl time.Duration) (bool, error) {
	payload, err := marshalWindowOps(ops)
	if err != nil {
		return false, AnnotateMarshalReviewCacheError(err)
	}
	ttlMs := int64(ttl / time.Millisecond)
	if ttlMs < 0 {
		ttlMs = 0
	}
	res, err := reviewWindowOpsAndPopScript.Run(ctx, r.rdb, []string{queueKey, leaseKey}, token, ownerPrefix, reviewID, ttlMs, string(payload)).Int64()
	if err != nil {
		return false, AnnotateApplyReviewWindowOpsError(ownerPrefix, err)
	}
	return res == 1, nil
}

// CheckDurability verifies that Redis persistence is enabled for the queue
// durability boundary.
func (r *repo) CheckDurability(ctx context.Context) (bool, error) {
	appendOnlyValues, err := r.rdb.ConfigGet(ctx, "appendonly").Result()
	if err != nil {
		return false, AnnotateDurabilityReviewError(err)
	}
	appendOnly := false
	if value, ok := appendOnlyValues["appendonly"]; ok {
		appendOnly = strings.EqualFold(value, "yes")
	}

	saveValues, err := r.rdb.ConfigGet(ctx, "save").Result()
	if err != nil {
		return false, AnnotateDurabilityReviewError(err)
	}
	snapshot := false
	if value, ok := saveValues["save"]; ok {
		snapshot = strings.TrimSpace(value) != ""
	}
	return appendOnly || snapshot, nil
}

func marshalWindowOps(ops []domain.WindowOp) ([]byte, error) {
	payload := make([]windowOpPayload, 0, len(ops))
	for _, op := range ops {
		item := windowOpPayload{Window: op.Window, Insert: make([]windowMemberPayload, 0, len(op.Insert)), Evict: make([]windowMemberPayload, 0, len(op.Evict))}
		for _, insert := range op.Insert {
			member, err := reviewMemberPayload(insert)
			if err != nil {
				return nil, err
			}
			item.Insert = append(item.Insert, member)
		}
		for _, evict := range op.Evict {
			member, err := reviewMemberPayload(evict)
			if err != nil {
				return nil, err
			}
			item.Evict = append(item.Evict, member)
		}
		payload = append(payload, item)
	}
	return json.Marshal(payload)
}

func reviewMemberPayload(member domain.ReviewMember) (windowMemberPayload, error) {
	if member.Review == nil {
		return windowMemberPayload{}, ErrNilReview
	}
	cache := mapper.MapDomainReviewToCache(member.Review)
	raw, err := json.Marshal(cache)
	if err != nil {
		return windowMemberPayload{}, err
	}
	return windowMemberPayload{Member: string(raw), Score: member.Score}, nil
}

func toInt(raw interface{}) (int, error) {
	switch v := raw.(type) {
	case int64:
		return int(v), nil
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, ErrInvalidReviewMember
	}
}

func toScore(raw interface{}) float64 {
	switch v := raw.(type) {
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case float64:
		return v
	case string:
		parsed, _ := strconv.ParseFloat(v, 64)
		return parsed
	default:
		return 0
	}
}

func toReviewFromCache(member interface{}, score interface{}) (*domain.Review, error) {
	raw, ok := member.(string)
	if !ok {
		return nil, ErrInvalidReviewMember
	}
	if strings.TrimSpace(raw) == "" {
		return nil, ErrInvalidReviewMember
	}
	review, err := MapZSetMemberToDomain(raw, toScore(score))
	if err != nil {
		return nil, err
	}
	return review, nil
}

type windowOpPayload struct {
	Window int
	Insert []windowMemberPayload
	Evict  []windowMemberPayload
}

type windowMemberPayload struct {
	Member string
	Score  float64
}

var reviewWindowEnqueueScript = redis.NewScript(`
local queueKey = KEYS[1]
local activeKey = KEYS[2]
local payload = ARGV[1]
local depth = redis.call("RPUSH", queueKey, payload)
redis.call("SADD", activeKey, queueKey)
return depth
`)

var reviewWindowSnapshotScript = redis.NewScript(`
local prefix = ARGV[1]
local maxWindows = 1024
local result = {}
for i = 0, maxWindows - 1 do
  local key = prefix .. ":window:" .. i
  local count = redis.call("ZCARD", key)
  if count == 0 then
    break
  end
  local newest = redis.call("ZREVRANGE", key, 0, 0, "WITHSCORES")
  local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
  table.insert(result, i)
  table.insert(result, count)
  table.insert(result, newest[1] or "")
  table.insert(result, newest[2] or "")
  table.insert(result, oldest[1] or "")
  table.insert(result, oldest[2] or "")
end
return result
`)

var reviewWindowOpsAndPopScript = redis.NewScript(`
local queueKey = KEYS[1]
local leaseKey = KEYS[2]
local token = ARGV[1]
local prefix = ARGV[2]
local reviewID = ARGV[3]
local ttlMs = tonumber(ARGV[4]) or 0
local ops = cjson.decode(ARGV[5])

if redis.call("GET", leaseKey) ~= token then
  return redis.error_reply("lease lost")
end

local rawHead = redis.call("LINDEX", queueKey, 0)
if not rawHead then
  return 0
end
local head = cjson.decode(rawHead)
if type(head) ~= "table" or head.review_id ~= reviewID then
  return 0
end

for _, op in ipairs(ops) do
  if type(op.Window) ~= "number" then
    return redis.error_reply("invalid window op")
  end
  if type(op.Insert) ~= "table" or type(op.Evict) ~= "table" then
    return redis.error_reply("invalid window op members")
  end
  for _, item in ipairs(op.Insert) do
    if type(item.Member) ~= "string" then
      return redis.error_reply("invalid insert member")
    end
    local score = tonumber(item.Score)
    if not score then
      return redis.error_reply("invalid insert score")
    end
  end
  for _, item in ipairs(op.Evict) do
    if type(item.Member) ~= "string" then
      return redis.error_reply("invalid evict member")
    end
  end
end

for _, op in ipairs(ops) do
  local key = prefix .. ":window:" .. op.Window
  for _, item in ipairs(op.Insert) do
    redis.call("ZADD", key, tonumber(item.Score), item.Member)
  end
  for _, item in ipairs(op.Evict) do
    redis.call("ZREM", key, item.Member)
  end
  local count = redis.call("ZCARD", key)
  if count == 0 then
    redis.call("DEL", key)
  elseif ttlMs > 0 and op.Window > 0 then
    redis.call("PEXPIRE", key, ttlMs)
  end
end

redis.call("LPOP", queueKey)
return 1
`)

var reviewWindowLeaseRenewScript = redis.NewScript(`
local leaseKey = KEYS[1]
local token = ARGV[1]
local ttlMs = tonumber(ARGV[2]) or 0
if redis.call("GET", leaseKey) == token then
  redis.call("PEXPIRE", leaseKey, ttlMs)
  return 1
end
return 0
`)

var reviewWindowLeaseReleaseScript = redis.NewScript(`
local leaseKey = KEYS[1]
local token = ARGV[1]
if redis.call("GET", leaseKey) == token then
  redis.call("DEL", leaseKey)
  return 1
end
return 0
`)

var reviewWindowInactiveCleanupScript = redis.NewScript(`
local queueKey = KEYS[1]
local activeKey = KEYS[2]
if redis.call("LLEN", queueKey) == 0 then
  redis.call("SREM", activeKey, queueKey)
  return 1
end
return 0
`)
