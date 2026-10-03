package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"review-service/config"
	app "review-service/internal/application"
	domain "review-service/internal/domain"
	eventbroker "review-service/internal/presentation/event_broker"
)

type projectionSvcStub struct {
	projected          *app.ProjectionResult
	projectErr         error
	projectCalls       int
	gigSummaryCalls    int
	sellerSummaryCalls int
	gigSummary         *domain.RatingSummary
	sellerSummary      *domain.RatingSummary
}

func (s *projectionSvcStub) CreateReview(context.Context, app.CreateReviewCommand) (*domain.Review, error) {
	return nil, errors.New("unexpected call")
}
func (s *projectionSvcStub) ListGigReviews(context.Context, app.ListGigReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (s *projectionSvcStub) ListSellerReviews(context.Context, app.ListSellerReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (s *projectionSvcStub) GetReviewsBySellerUsername(context.Context, string, string) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (s *projectionSvcStub) GetGigRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	s.gigSummaryCalls++
	if s.gigSummary != nil {
		return s.gigSummary, nil
	}
	return nil, domain.ErrReviewNotFound
}
func (s *projectionSvcStub) GetUserRatingSummaryByUsername(context.Context, string) (*domain.RatingSummary, error) {
	s.sellerSummaryCalls++
	if s.sellerSummary != nil {
		return s.sellerSummary, nil
	}
	return nil, domain.ErrReviewNotFound
}
func (s *projectionSvcStub) ProjectReview(context.Context, *domain.Review, app.ProjectionPolicy) (*app.ProjectionResult, error) {
	s.projectCalls++
	if s.projectErr != nil {
		return nil, s.projectErr
	}
	return s.projected, nil
}

type projectionReadStub struct {
	gigSummary    *domain.RatingSummary
	sellerSummary *domain.RatingSummary
	upsertGig     int
	upsertSeller  int
}

func (r *projectionReadStub) UpsertGigRating(context.Context, string, int32) error {
	r.upsertGig++
	return nil
}
func (r *projectionReadStub) UpsertSellerRating(context.Context, string, int32) error {
	r.upsertSeller++
	return nil
}
func (r *projectionReadStub) UpsertSellerRatingByUsername(context.Context, string, int32) error {
	r.upsertSeller++
	return nil
}
func (r *projectionReadStub) GetGigRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	if r.gigSummary != nil {
		return r.gigSummary, nil
	}
	return nil, domain.ErrReviewNotFound
}
func (r *projectionReadStub) GetSellerRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	if r.sellerSummary != nil {
		return r.sellerSummary, nil
	}
	return nil, domain.ErrReviewNotFound
}
func (r *projectionReadStub) GetSellerRatingSummaryByUsername(context.Context, string) (*domain.RatingSummary, error) {
	if r.sellerSummary != nil {
		return r.sellerSummary, nil
	}
	return nil, domain.ErrReviewNotFound
}
func (r *projectionReadStub) GetSellerIDByUsername(context.Context, string) (string, error) {
	return "", domain.ErrReviewNotFound
}
func (r *projectionReadStub) SetGigRating(context.Context, string, domain.RatingSummary) error {
	return nil
}
func (r *projectionReadStub) SetSellerRating(context.Context, string, domain.RatingSummary) error {
	return nil
}
func (r *projectionReadStub) SetSellerRatingByUsername(context.Context, string, domain.RatingSummary) error {
	return nil
}
func (r *projectionReadStub) SetSellerIDByUsername(context.Context, string, string) error {
	return nil
}
func (r *projectionReadStub) ListByGigID(context.Context, domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (r *projectionReadStub) ListBySellerID(context.Context, domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (r *projectionReadStub) ListByGigWindow(context.Context, string, int) (*domain.ListReviewsResult, error) {
	return nil, domain.ErrReviewNotFound
}
func (r *projectionReadStub) ListBySellerWindow(context.Context, string, int) (*domain.ListReviewsResult, error) {
	return nil, domain.ErrReviewNotFound
}
func (r *projectionReadStub) UpsertGigWindow(context.Context, string, int, []*domain.Review, bool, time.Duration) error {
	return nil
}
func (r *projectionReadStub) UpsertSellerWindow(context.Context, string, int, []*domain.Review, bool, time.Duration) error {
	return nil
}
func (r *projectionReadStub) DeleteByID(context.Context, string) error { return nil }

type projectionCoordStub struct {
	gigCalls    int
	sellerCalls int
}

func (c *projectionCoordStub) Start(context.Context) error { return nil }
func (c *projectionCoordStub) EnqueueGigReview(context.Context, *domain.Review) error {
	c.gigCalls++
	return nil
}
func (c *projectionCoordStub) EnqueueSellerReview(context.Context, *domain.Review) error {
	c.sellerCalls++
	return nil
}

type projectionUserStub struct {
	preview *app.UserPreview
	err     error
}

func (u *projectionUserStub) GetUserPreviewByID(context.Context, string) (*app.UserPreview, error) {
	if u.preview != nil {
		return u.preview, u.err
	}
	return nil, u.err
}

func (u *projectionUserStub) GetUserPreviewByIDNoCache(context.Context, string) (*app.UserPreview, error) {
	if u.preview != nil {
		return u.preview, u.err
	}
	return nil, u.err
}

func (u *projectionUserStub) GetDetailedUserByUsername(context.Context, string) (*app.UserPreview, error) {
	return u.preview, u.err
}

type projectionBrokerStub struct {
	published []string
	payloads  [][]byte
}

func (b *projectionBrokerStub) Publish(_ context.Context, subject string, payload []byte) error {
	b.published = append(b.published, subject)
	b.payloads = append(b.payloads, append([]byte(nil), payload...))
	return nil
}
func (b *projectionBrokerStub) RunPullConsumer(context.Context, config.PullConsumerConfig, eventbroker.MessageHandler) error {
	return nil
}
func (b *projectionBrokerStub) Close() {}

type testLogger struct{}

func (testLogger) Debug(string, ...logging.Field)       {}
func (testLogger) Info(string, ...logging.Field)        {}
func (testLogger) Warn(string, ...logging.Field)        {}
func (testLogger) Error(string, ...logging.Field)       {}
func (testLogger) With(...logging.Field) logging.Logger { return testLogger{} }
func (testLogger) Sync() error                          { return nil }

func TestHandleGigProjectionRequestedEnqueuesAndRetries(t *testing.T) {
	svc := &projectionSvcStub{projected: &app.ProjectionResult{Review: &domain.Review{ID: "review-1"}, AuthorRetry: true, AvatarRetry: true}}
	coord := &projectionCoordStub{}
	broker := &projectionBrokerStub{}
	s := &ReviewProjectionSubscriber{
		broker: broker,
		svc:    svc,
		read:   &projectionReadStub{},
		coord:  coord,
		users:  &projectionUserStub{preview: &app.UserPreview{UserID: "seller-1", Username: "seller-username"}},
		cfg:    &config.Config{NATS: config.NATSConfig{ReviewAuthorRetrySubject: "author.retry", ReviewAvatarRetrySubject: "avatar.retry"}},
		log:    testLogger{},
	}

	if err := s.HandleGigProjectionRequested(context.Background(), &domain.Review{ID: "review-1"}); err != nil {
		t.Fatalf("HandleGigProjectionRequested returned error: %v", err)
	}
	if svc.projectCalls != 1 || coord.gigCalls != 1 {
		t.Fatalf("expected projector and coordinator calls, got %+v %+v", svc, coord)
	}
	if len(broker.published) != 2 {
		t.Fatalf("expected retry publishes, got %#v", broker.published)
	}
}

func TestHandleUserProjectionRequestedAndRetryShortCircuits(t *testing.T) {
	svc := &projectionSvcStub{projected: &app.ProjectionResult{Review: &domain.Review{ID: "review-2"}}}
	coord := &projectionCoordStub{}
	s := &ReviewProjectionSubscriber{
		broker: &projectionBrokerStub{},
		svc:    svc,
		read:   &projectionReadStub{},
		coord:  coord,
		users:  &projectionUserStub{preview: &app.UserPreview{UserID: "seller-1", Username: "seller-username"}},
		cfg:    &config.Config{},
		log:    testLogger{},
	}

	if err := s.HandleUserProjectionRequested(context.Background(), &domain.Review{ID: "review-2"}); err != nil {
		t.Fatalf("HandleUserProjectionRequested returned error: %v", err)
	}
	if coord.sellerCalls != 1 {
		t.Fatalf("expected seller queue enqueue")
	}

	if err := s.HandleAuthorRetryRequested(context.Background(), &domain.Review{ID: "review-3"}); err != nil {
		t.Fatalf("HandleAuthorRetryRequested returned error: %v", err)
	}
	if err := s.HandleAvatarRetryRequested(context.Background(), &domain.Review{ID: "review-4"}); err != nil {
		t.Fatalf("HandleAvatarRetryRequested returned error: %v", err)
	}
	if svc.projectCalls != 3 {
		t.Fatalf("expected retry handlers to invoke projector, got %d", svc.projectCalls)
	}
}

func TestWriteGigAndSellerRating(t *testing.T) {
	read := &projectionReadStub{gigSummary: &domain.RatingSummary{TotalReviews: 5}, sellerSummary: &domain.RatingSummary{TotalReviews: 7}}
	svc := &projectionSvcStub{gigSummary: &domain.RatingSummary{TotalReviews: 9}, sellerSummary: &domain.RatingSummary{TotalReviews: 11}}
	s := &ReviewProjectionSubscriber{
		broker: &projectionBrokerStub{},
		svc:    svc,
		read:   read,
		coord:  &projectionCoordStub{},
		users:  &projectionUserStub{preview: &app.UserPreview{UserID: "seller-1", Username: "seller-username"}},
		cfg:    &config.Config{},
		log:    testLogger{},
	}

	if err := s.WriteGigRating(context.Background(), &domain.Review{GigID: "gig-1", Rating: 5}); err != nil {
		t.Fatalf("WriteGigRating returned error: %v", err)
	}
	if err := s.WriteSellerRating(context.Background(), &domain.Review{SellerID: "seller-1", Rating: 4}); err != nil {
		t.Fatalf("WriteSellerRating returned error: %v", err)
	}
	if read.upsertGig != 1 || read.upsertSeller != 1 {
		t.Fatalf("expected cache updates on hit, got %+v", read)
	}

	read = &projectionReadStub{}
	s.read = read
	if err := s.WriteGigRating(context.Background(), &domain.Review{GigID: "gig-2", Rating: 5}); err != nil {
		t.Fatalf("WriteGigRating miss returned error: %v", err)
	}
	if err := s.WriteSellerRating(context.Background(), &domain.Review{SellerID: "seller-1", Rating: 4}); err != nil {
		t.Fatalf("WriteSellerRating miss returned error: %v", err)
	}
	if svc.gigSummaryCalls != 1 || svc.sellerSummaryCalls != 1 {
		t.Fatalf("expected summary lookups on miss, got gig=%d seller=%d", svc.gigSummaryCalls, svc.sellerSummaryCalls)
	}
}
