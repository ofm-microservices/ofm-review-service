package nats

import (
	"time"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"review-service/config"

	"github.com/nats-io/nats.go"
)

type bootstrapConn interface {
	JetStream() (jetStreamManager, error)
	Close()
}

type jetStreamManager interface {
	AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error)
	UpdateStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error)
}

type realBootstrapConn struct {
	*nats.Conn
}

func (c realBootstrapConn) JetStream() (jetStreamManager, error) {
	return c.Conn.JetStream()
}

var connectBootstrap = func(cfg config.NATSConfig) (bootstrapConn, error) {
	nc, err := Connect(cfg)
	if err != nil {
		return nil, err
	}
	return realBootstrapConn{Conn: nc}, nil
}

// EnsureStream creates or updates the JetStream stream required by review-service.
func EnsureStream(cfg config.NATSConfig, log logging.Logger) error {
	if log == nil {
		return ErrNilLogger
	}

	lg := log.With(logging.String("module", "jetstream-bootstrap"))
	lg.Info("ensuring jetstream stream",
		logging.String("stream", cfg.ReviewEventsStream),
		logging.String("gig_projection_subject", cfg.ReviewGigProjectionSubject),
		logging.String("user_projection_subject", cfg.ReviewUserProjectionSubject),
		logging.String("gig_rating_subject", cfg.ReviewGigRatingSubject),
		logging.String("seller_rating_subject", cfg.ReviewSellerRatingSubject),
	)

	nc, err := connectBootstrap(cfg)
	if err != nil {
		return err
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		return AnnotateInitJetStreamContextError(err)
	}

	streamCfg := &nats.StreamConfig{
		Name:      cfg.ReviewEventsStream,
		Subjects:  []string{cfg.ReviewGigProjectionSubject, cfg.ReviewUserProjectionSubject, cfg.ReviewGigRatingSubject, cfg.ReviewSellerRatingSubject, cfg.ReviewAuthorRetrySubject, cfg.ReviewAvatarRetrySubject},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		Replicas:  1,
		MaxAge:    7 * 24 * time.Hour,
	}

	_, err = js.AddStream(streamCfg)
	if err != nil {
		if _, updateErr := js.UpdateStream(streamCfg); updateErr != nil {
			return AnnotateEnsureStreamError(streamCfg.Name, err, updateErr)
		}
	}

	lg.Info("jetstream stream ensured",
		logging.String("review_events_stream", streamCfg.Name),
	)
	return nil
}

// Connect establishes the low-level NATS connection used by bootstrap code.
func Connect(cfg config.NATSConfig) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name("review-service"),
		nats.MaxReconnects(-1),
	}

	if cfg.Review != "" {
		opts = append(opts, nats.UserInfo(cfg.Review, cfg.Password))
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, AnnotateConnectToNATSError(err)
	}

	return nc, nil
}
