package tmp

import (
	"context"
	"errors"
	indexer "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/client"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/integrations/keybase"
	v2 "github.com/babylonlabs-io/staking-api-service/internal/v2/db/client"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	lock "github.com/square/mongo-lock"
	"time"
)

const (
	resourceName = "fp.fetch_and_save_logo"

	lockDuration = 5 * time.Minute
)

type Service struct {
	indexerDB indexer.IndexerDBClient
	v2DB      v2.V2DBClient

	lockClient    *lock.Client
	keybaseClient *keybase.Client
}

func New(indexerDB indexer.IndexerDBClient, v2DB v2.V2DBClient, lockClient *lock.Client, keybaseClient *keybase.Client) *Service {
	return &Service{
		indexerDB:     indexerDB,
		v2DB:          v2DB,
		lockClient:    lockClient,
		keybaseClient: keybaseClient,
	}
}

func (s *Service) Run() {
	lockID, err := s.lock()
	if err != nil {
		if errors.Is(err, lock.ErrAlreadyLocked) {
			log.Info().Msgf("Lock on %s already acquired", resourceName)
		} else {
			log.Err(err).Msgf("Failed to lock resource %s", resourceName)
		}
		return
	}
	defer s.unlock(lockID)

	ctx := context.TODO()
	log := log.Ctx(ctx)

	existingLogos, err := s.getExistingLogos(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to get existing logos")
		return
	}

	fps, err := s.indexerDB.GetFinalityProviders(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to get finality providers")
		return
	}

	for _, fp := range fps {
		identity := fp.Description.Identity

		// if logo for given identity already exists - skip
		if _, exists := existingLogos[identity]; exists {
			continue
		}

		logoURL, err := s.keybaseClient.GetLogoURL(ctx, identity)
		if err != nil {
			log.Err(err).Msgf("Failed to get logo URL for %s", identity)
			continue
		}

		err = s.v2DB.InsertFinalityProviderLogo(ctx, identity, logoURL)
		if err != nil {
			log.Err(err).
				Str("identity", identity).
				Str("logoURL", logoURL).
				Msg("Failed to insert finality provider logo")
		}
	}
}

func (s *Service) getExistingLogos(ctx context.Context) (map[string]struct{}, error) {
	logos, err := s.v2DB.GetFinalityProviderLogos(ctx)
	if err != nil {
		return nil, err
	}

	set := map[string]struct{}{}
	for _, logo := range logos {
		set[logo.Id] = struct{}{}
	}
	return set, nil
}

func (s *Service) lock() (uuid.UUID, error) {
	ctx := context.TODO()
	lockID := uuid.New()
	err := s.lockClient.XLock(ctx, resourceName, lockID.String(), lock.LockDetails{
		TTL: uint(lockDuration.Seconds()),
	})

	return lockID, err
}

func (s *Service) unlock(lockID uuid.UUID) {
	ctx := context.TODO()
	_, err := s.lockClient.Unlock(ctx, lockID.String())
	if err != nil {
		log.Err(err).Msgf("Failed to unlock lock %s", lockID)
	}
}
