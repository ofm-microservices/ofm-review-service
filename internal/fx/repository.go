package appfx

import (
	domain "review-service/internal/domain"
	readrepo "review-service/internal/infra/read/redis"
	writerepo "review-service/internal/infra/write/yugabyte"

	"github.com/jmoiron/sqlx"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// RepoModule wires write- and read-model repositories into the FX graph.
var RepoModule = fx.Options(
	fx.Provide(
		writerepo.NewPgErrorTranslator,
		ProvideWriteRepo,
		ProvideReadRepo,
	),
)

// ProvideWriteRepo constructs the Yugabyte-backed review repository.
func ProvideWriteRepo(dbx *sqlx.DB, translator writerepo.DBErrorTranslator, lg logging.Logger) (domain.ReviewRepository, error) {
	return writerepo.New(dbx, translator, lg)
}

// ProvideReadRepo constructs the Redis-backed review read repository.
func ProvideReadRepo(rdb *redis.Client, lg logging.Logger) (domain.ReviewReadRepository, domain.ReviewProjectionRepository, error) {
	repo, err := readrepo.New(rdb, lg)
	if err != nil {
		return nil, nil, err
	}
	projectionRepo, ok := repo.(domain.ReviewProjectionRepository)
	if !ok {
		return nil, nil, readrepo.ErrNilRedisClient
	}
	return repo, projectionRepo, nil
}
