package api

import (
	_ "github.com/babylonlabs-io/staking-api-service/docs"
	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (a *Server) SetupRoutes(r *chi.Mux) {
	handlers := a.handlers
	// Common routes
	r.Get("/healthcheck", registerHandler(handlers.SharedHandler.HealthCheck))
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/v1/staker/pubkey-lookup", registerHandler(handlers.V1Handler.GetPubKeys))
	r.Get("/v1/staker/delegation/check", registerHandler(handlers.V1Handler.CheckStakerDelegationExist))

	// Only register these routes if the asset has been configured
	// The endpoints are used to check ordinals within the UTXOs
	// Don't deprecate this endpoint
	if a.cfg.Assets != nil {
		r.Post("/v1/ordinals/verify-utxos", registerHandler(handlers.SharedHandler.VerifyUTXOs))
	}

	// V2 API
	r.Get("/v2/network-info", registerHandler(handlers.V2Handler.GetNetworkInfo))
	r.Get("/v2/finality-providers", registerHandler(handlers.V2Handler.GetFinalityProviders))
	r.Get("/v2/delegation", registerHandler(handlers.V2Handler.GetDelegation))
	r.Get("/v2/delegations", registerHandler(handlers.V2Handler.GetDelegations))
	r.Get("/v2/stats", registerHandler(handlers.V2Handler.GetOverallStats))
	r.Get("/v2/staker/stats", registerHandler(handlers.V2Handler.GetStakerStats))
	r.Get("/v2/prices", registerHandler(handlers.V2Handler.GetPrices))

	// Legacy endpoints needed to support phase-1 delegations to unbond.
	// These will be deprecated once all phase-1 delegations are either withdrawn or registered into phase-2.
	r.Post("/v1/unbonding", registerHandler(handlers.V1Handler.UnbondDelegation))
	r.Get("/v1/unbonding/eligibility", registerHandler(handlers.V1Handler.GetUnbondingEligibility))
	r.Get("/v1/staker/delegations", registerHandler(handlers.V1Handler.GetStakerDelegations))

	// Deprecated endpoints that were used in phase-1. Will be removed in the future
	r.Get("/v1/global-params", registerHandler(handlers.V1Handler.GetBabylonGlobalParams))
	r.Get("/v1/finality-providers", registerHandler(handlers.V1Handler.GetFinalityProviders))
	r.Get("/v1/stats", registerHandler(handlers.V1Handler.GetOverallStats))
	r.Get("/v1/stats/staker", registerHandler(handlers.V1Handler.GetStakersStats))
	r.Get("/v1/delegation", registerHandler(handlers.V1Handler.GetDelegationByTxHash))
}
