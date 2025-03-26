package v2service

import (
	"context"
	v2dbclient "github.com/babylonlabs-io/staking-api-service/internal/v2/db/client"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
	"sync"
	"time"
)

type overallStatsService struct {
	db v2dbclient.V2DBClient

	ttl       time.Duration
	mx        sync.RWMutex
	updatedAt time.Time
	value     *v2dbmodel.V2OverallStatsDocument
}

func newOverallStatsService(db v2dbclient.V2DBClient, ttl time.Duration) *overallStatsService {
	return &overallStatsService{
		db:  db,
		ttl: ttl,
		// zero values for other fields are fine
	}
}

func (s *overallStatsService) getOverallStatsFromDB(ctx context.Context) (*v2dbmodel.V2OverallStatsDocument, error) {
	s.mx.RLock()
	if s.value != nil && time.Since(s.updatedAt) < s.ttl {
		defer s.mx.RUnlock()
		return s.value, nil
	}
	s.mx.RUnlock()

	s.mx.Lock()
	defer s.mx.Unlock()
	stats, err := s.db.GetOverallStats(ctx)
	if err != nil {
		return nil, err
	}

	s.value = stats
	s.updatedAt = time.Now()

	return s.value, nil
}
