package business

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/repository"
)

// BuildInfo holds compile-time version metadata injected at bootstrap.
type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

// HealthService orchestrates health, readiness, and version checks.
type HealthService struct {
	repo      repository.HealthRepository
	buildInfo BuildInfo
}

// NewHealthService constructs a HealthService.
func NewHealthService(repo repository.HealthRepository, buildInfo BuildInfo) *HealthService {
	return &HealthService{repo: repo, buildInfo: buildInfo}
}

// Liveness returns process-up status without touching the database.
func (s *HealthService) Liveness() entity.Health {
	return entity.Health{Status: constant.StatusOK}
}

// Readiness pings the database with a short timeout.
func (s *HealthService) Readiness(ctx context.Context) (entity.Readiness, error) {
	pingCtx, cancel := context.WithTimeout(ctx, constant.ReadinessPingTimeout)
	defer cancel()

	if err := s.repo.Ping(pingCtx); err != nil {
		return entity.Readiness{
			Status:   constant.StatusUnavailable,
			Database: constant.DatabaseDown,
		}, fmt.Errorf("database ping: %w", err)
	}

	return entity.Readiness{
		Status:   constant.StatusOK,
		Database: constant.DatabaseUp,
	}, nil
}

// Version returns build metadata.
func (s *HealthService) Version() entity.Version {
	return entity.Version{
		Version:   s.buildInfo.Version,
		Commit:    s.buildInfo.Commit,
		BuildTime: s.buildInfo.BuildTime,
	}
}
