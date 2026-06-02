package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	domain "review-service/internal/domain"
)

const reviewQueueActiveKey = "reviews.queue.active"

var errStaleWindowJob = errors.New("stale review window job")

type ownerKind string

const (
	ownerKindGig    ownerKind = "gig"
	ownerKindSeller ownerKind = "seller"
)

type ownerRef struct {
	kind ownerKind
	id   string
}

func newGigOwner(id string) ownerRef {
	return ownerRef{kind: ownerKindGig, id: strings.TrimSpace(id)}
}

func newSellerOwner(id string) ownerRef {
	return ownerRef{kind: ownerKindSeller, id: strings.TrimSpace(id)}
}

func (o ownerRef) queueKey() string {
	return fmt.Sprintf("reviews.queue.%s.%s", o.kind, o.id)
}

func (o ownerRef) leaseKey() string {
	return o.queueKey() + ".lease"
}

func (o ownerRef) windowPrefix() string {
	return fmt.Sprintf("%s:reviews:%s", o.kind, o.id)
}

// reviewWindowCoordinator serializes review-window mutations through Redis
// queues and leases.
type reviewWindowCoordinator struct {
	repo      domain.ReviewProjectionRepository
	source    domain.ReviewRepository
	projector ReviewService
	paging    PaginationConfig
	cfg       WindowCoordinatorConfig
	log       Logger
	startMu   sync.Once
	startErr  error
	running   sync.Map
}

// NewReviewWindowCoordinator constructs the Redis-backed review-window
// coordinator.
func NewReviewWindowCoordinator(repo domain.ReviewProjectionRepository, source domain.ReviewRepository, projector ReviewService, paging PaginationConfig, cfg WindowCoordinatorConfig, log Logger) (ReviewWindowCoordinator, error) {
	if repo == nil {
		return nil, ErrNilReviewReadRepository
	}
	if source == nil {
		return nil, ErrNilReviewRepository
	}
	if projector == nil {
		return nil, ErrNilReviewService
	}
	if paging.WindowSize <= 0 {
		return nil, ErrInvalidPaginationConfig
	}
	if cfg.LeaseTTL <= 0 || cfg.LeaseRenewInterval <= 0 || cfg.ReconcileInterval <= 0 {
		return nil, ErrInvalidPaginationConfig
	}
	if log == nil {
		return nil, ErrNilLogger
	}

	return &reviewWindowCoordinator{
		repo:      repo,
		source:    source,
		projector: projector,
		paging:    paging,
		cfg:       cfg,
		log:       log.With(logging.String("module", "review-window-coordinator")),
	}, nil
}

// Start verifies queue durability and launches the reconciler loop.
func (c *reviewWindowCoordinator) Start(ctx context.Context) error {
	if c == nil {
		return ErrNilLogger
	}
	c.startMu.Do(func() {
		if c.cfg.AckAfterEnqueue {
			persisted, err := c.repo.CheckDurability(ctx)
			if err != nil {
				c.startErr = err
				return
			}
			if !persisted {
				c.startErr = ErrQueueDurabilityDisabled
				return
			}
		}
		go c.runReconciler(ctx)
	})
	return c.startErr
}

// EnqueueGigReview stores a gig review job durably in Redis and kicks best-effort
// drain processing.
func (c *reviewWindowCoordinator) EnqueueGigReview(ctx context.Context, review *domain.Review) error {
	if review == nil {
		return domain.ErrReviewNotFound
	}
	return c.enqueueAndKick(ctx, newGigOwner(review.GigID), review)
}

// EnqueueSellerReview stores a seller review job durably in Redis and kicks
// best-effort drain processing.
func (c *reviewWindowCoordinator) EnqueueSellerReview(ctx context.Context, review *domain.Review) error {
	if review == nil {
		return domain.ErrReviewNotFound
	}
	ownerID := review.SellerID
	if strings.TrimSpace(ownerID) == "" && review.Author != nil {
		ownerID = review.Author.UserID
	}
	return c.enqueueAndKick(ctx, newSellerOwner(ownerID), review)
}

func (c *reviewWindowCoordinator) enqueueAndKick(ctx context.Context, owner ownerRef, review *domain.Review) error {
	owner.id = strings.TrimSpace(owner.id)
	if owner.id == "" {
		return domain.ErrReviewNotFound
	}

	depth, err := c.repo.EnqueueWindowJob(ctx, owner.queueKey(), reviewQueueActiveKey, review)
	if err != nil {
		return err
	}
	if c.cfg.QueueDepthWarnThreshold > 0 && int(depth) > c.cfg.QueueDepthWarnThreshold {
		c.log.Warn("review window queue depth exceeded threshold",
			logging.Operation("review.window.queue_depth"),
			logging.String("queue", owner.queueKey()),
			logging.Int("depth", int(depth)),
			logging.Int("threshold", c.cfg.QueueDepthWarnThreshold),
		)
	}

	go c.kickDrain(context.WithoutCancel(ctx), owner)
	return nil
}

func (c *reviewWindowCoordinator) kickDrain(ctx context.Context, owner ownerRef) {
	queueKey := owner.queueKey()
	if _, running := c.running.LoadOrStore(queueKey, struct{}{}); running {
		return
	}

	go func() {
		defer c.running.Delete(queueKey)
		if err := c.drainOwner(ctx, owner); err != nil && !errors.Is(err, domain.ErrReviewNotFound) && !errors.Is(err, errStaleWindowJob) {
			c.log.Error("review window drain failed",
				logging.Operation("review.window.drain"),
				logging.String("queue", queueKey),
				logging.String("lease", owner.leaseKey()),
				logging.Err(err),
			)
		}
	}()
}

func (c *reviewWindowCoordinator) drainOwner(ctx context.Context, owner ownerRef) error {
	token := uuid.NewString()
	ok, err := c.repo.TryAcquireLease(ctx, owner.leaseKey(), token, c.cfg.LeaseTTL)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	renewDone := make(chan struct{})
	renewErr := make(chan error, 1)
	go c.renewLeaseLoop(ctx, owner.leaseKey(), token, renewDone, renewErr)
	defer close(renewDone)
	defer func() {
		if released, relErr := c.repo.ReleaseLease(context.WithoutCancel(ctx), owner.leaseKey(), token); relErr != nil {
			c.log.Error("review window lease release failed",
				logging.Operation("review.window.lease_release"),
				logging.String("lease", owner.leaseKey()),
				logging.Err(relErr),
			)
		} else if !released {
			c.log.Warn("review window lease release lost ownership",
				logging.Operation("review.window.lease_release"),
				logging.String("lease", owner.leaseKey()),
			)
		}
	}()

	queueKey := owner.queueKey()
	windowPrefix := owner.windowPrefix()
	for {
		select {
		case renewErr := <-renewErr:
			if renewErr != nil {
				return renewErr
			}
		default:
		}

		review, peekErr := c.repo.PeekWindowJob(ctx, queueKey)
		if peekErr != nil {
			if errors.Is(peekErr, domain.ErrReviewNotFound) {
				break
			}
			return peekErr
		}
		if review == nil {
			break
		}

		snapshots, snapErr := c.repo.LoadWindowSnapshots(ctx, windowPrefix)
		if snapErr != nil {
			return snapErr
		}

		if len(snapshots) == 0 {
			seedOps, seedReviewID, seedErr := c.seedOwnerWindow(ctx, owner)
			if seedErr != nil {
				return seedErr
			}
			handled, popErr := c.repo.ApplyWindowOpsAndPopJob(ctx, queueKey, owner.leaseKey(), token, windowPrefix, seedReviewID, seedOps, c.paging.WindowTTL)
			if popErr != nil {
				return popErr
			}
			if !handled {
				return nil
			}
			continue
		}

		ops, ok := planWindowOps(review, snapshots, c.paging.WindowSize)
		if !ok {
			handled, popErr := c.repo.ApplyWindowOpsAndPopJob(ctx, queueKey, owner.leaseKey(), token, windowPrefix, review.ID, nil, c.paging.WindowTTL)
			if popErr != nil {
				return popErr
			}
			if handled {
				c.log.Warn("dropping stale review window job",
					logging.Operation("review.window.drop"),
					logging.String("queue", queueKey),
					logging.String("review_id", review.ID),
				)
			}
			continue
		}

		handled, err := c.repo.ApplyWindowOpsAndPopJob(ctx, queueKey, owner.leaseKey(), token, windowPrefix, review.ID, ops, c.paging.WindowTTL)
		if err != nil {
			return err
		}
		if !handled {
			return nil
		}
	}

	if _, err := c.repo.RemoveActiveQueueIfEmpty(ctx, reviewQueueActiveKey, queueKey); err != nil {
		return err
	}
	return nil
}

func (c *reviewWindowCoordinator) renewLeaseLoop(ctx context.Context, leaseKey, token string, done <-chan struct{}, errs chan<- error) {
	ticker := time.NewTicker(c.cfg.LeaseRenewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			ok, err := c.repo.RenewLease(context.WithoutCancel(ctx), leaseKey, token, c.cfg.LeaseTTL)
			if err != nil {
				select {
				case errs <- err:
				default:
				}
				return
			}
			if !ok {
				select {
				case errs <- ErrQueueLeaseLost:
				default:
				}
				return
			}
		}
	}
}

func (c *reviewWindowCoordinator) runReconciler(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.ReconcileInterval)
	defer ticker.Stop()

	c.reconcileQueues(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.reconcileQueues(ctx)
		}
	}
}

func (c *reviewWindowCoordinator) reconcileQueues(ctx context.Context) {
	queues, err := c.repo.ActiveQueueKeys(ctx, reviewQueueActiveKey)
	if err != nil {
		c.log.Error("review window reconcile failed to list active queues",
			logging.Operation("review.window.reconcile"),
			logging.Err(err),
		)
		return
	}

	for _, queueKey := range queues {
		queueKey = strings.TrimSpace(queueKey)
		if queueKey == "" {
			continue
		}
		owner, ok := parseOwnerFromQueueKey(queueKey)
		if !ok {
			c.log.Warn("review window reconcile skipped unknown queue key",
				logging.Operation("review.window.reconcile"),
				logging.String("queue", queueKey),
			)
			continue
		}

		depth, err := c.repo.QueueLength(ctx, queueKey)
		if err != nil {
			c.log.Error("review window reconcile queue depth failed",
				logging.Operation("review.window.reconcile"),
				logging.String("queue", queueKey),
				logging.Err(err),
			)
			continue
		}
		if depth == 0 {
			if _, err := c.repo.RemoveActiveQueueIfEmpty(ctx, reviewQueueActiveKey, queueKey); err != nil {
				c.log.Error("review window reconcile inactive cleanup failed",
					logging.Operation("review.window.reconcile"),
					logging.String("queue", queueKey),
					logging.Err(err),
				)
			}
			continue
		}

		go c.kickDrain(context.WithoutCancel(ctx), owner)
	}
}

func (c *reviewWindowCoordinator) seedOwnerWindow(ctx context.Context, owner ownerRef) ([]domain.WindowOp, string, error) {
	limit := c.paging.WindowSize
	if limit <= 0 {
		limit = 1
	}

	var (
		result *domain.ListReviewsResult
		err    error
	)
	switch owner.kind {
	case ownerKindGig:
		result, err = c.source.ListByGigID(ctx, domain.ListReviewsQuery{GigID: owner.id, Limit: limit})
	case ownerKindSeller:
		result, err = c.source.ListBySellerID(ctx, domain.ListReviewsQuery{SellerID: owner.id, Limit: limit})
	default:
		return nil, "", domain.ErrReviewNotFound
	}
	if err != nil {
		return nil, "", err
	}
	if result == nil || len(result.Reviews) == 0 {
		return nil, "", domain.ErrReviewNotFound
	}

	members := make([]domain.ReviewMember, 0, len(result.Reviews))
	for _, review := range result.Reviews {
		if review == nil {
			continue
		}
		projected, err := c.projector.ProjectReview(ctx, review, ProjectionPolicy{EnrichAuthor: true, EnrichAvatar: true})
		if err != nil {
			return nil, "", err
		}
		if projected == nil || projected.Review == nil {
			continue
		}
		members = append(members, domain.ReviewMember{Review: projected.Review, Score: float64(projected.Review.CreatedAt.UTC().UnixMicro())})
	}
	if len(members) == 0 {
		return nil, "", domain.ErrReviewNotFound
	}

	return []domain.WindowOp{{Window: 0, Insert: members}}, result.Reviews[0].ID, nil
}

func parseOwnerFromQueueKey(queueKey string) (ownerRef, bool) {
	parts := strings.Split(queueKey, ".")
	if len(parts) < 4 {
		return ownerRef{}, false
	}
	if parts[0] != "reviews" || parts[1] != "queue" {
		return ownerRef{}, false
	}
	kind := ownerKind(parts[2])
	id := strings.Join(parts[3:], ".")
	if id == "" {
		return ownerRef{}, false
	}
	switch kind {
	case ownerKindGig, ownerKindSeller:
		return ownerRef{kind: kind, id: id}, true
	default:
		return ownerRef{}, false
	}
}

func planWindowOps(review *domain.Review, snapshots []domain.WindowSnapshot, windowSize int) ([]domain.WindowOp, bool) {
	if review == nil || windowSize <= 0 {
		return nil, false
	}
	member := domain.ReviewMember{Review: review, Score: float64(review.CreatedAt.UTC().UnixMicro())}
	if len(snapshots) == 0 {
		return []domain.WindowOp{{Window: 0, Insert: []domain.ReviewMember{member}}}, true
	}

	target := -1
	for _, snap := range snapshots {
		if review.CreatedAt.UTC().After(snap.Oldest.Review.CreatedAt.UTC()) || review.CreatedAt.Equal(snap.Oldest.Review.CreatedAt) {
			target = snap.Window
			break
		}
	}
	if target < 0 {
		if review.CreatedAt.UTC().After(snapshots[0].Newest.Review.CreatedAt.UTC()) || review.CreatedAt.Equal(snapshots[0].Newest.Review.CreatedAt) {
			target = 0
		} else {
			return nil, false
		}
	}

	ops := make([]domain.WindowOp, 0, 2)
	carry := member
	window := target
	for {
		snap, ok := windowSnapshotByIndex(snapshots, window)
		if !ok {
			if len(ops) == 0 {
				return nil, false
			}
			ops = append(ops, domain.WindowOp{Window: window, Insert: []domain.ReviewMember{carry}})
			return ops, true
		}

		op := domain.WindowOp{Window: window, Insert: []domain.ReviewMember{carry}}
		if snap.Count >= windowSize {
			op.Evict = []domain.ReviewMember{snap.Oldest}
			carry = snap.Oldest
			ops = append(ops, op)
			window++
			continue
		}

		ops = append(ops, op)
		return ops, true
	}
}

func windowSnapshotByIndex(snapshots []domain.WindowSnapshot, window int) (domain.WindowSnapshot, bool) {
	for _, snap := range snapshots {
		if snap.Window == window {
			return snap, true
		}
	}
	return domain.WindowSnapshot{}, false
}
