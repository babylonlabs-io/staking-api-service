package api

import (
	_ "github.com/babylonlabs-io/staking-api-service/docs"
	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (a *Server) SetupRoutes(r *chi.Mux) {
	handlers := a.handlers
	// Extend on the healthcheck endpoint here
	r.Get("/healthcheck", registerHandler(handlers.V1.HealthCheck))

	r.Get("/v1/staker/delegations", registerHandler(handlers.V1.GetStakerDelegations))
	r.Post("/v1/unbonding", registerHandler(handlers.V1.UnbondDelegation))
	r.Get("/v1/unbonding/eligibility", registerHandler(handlers.V1.GetUnbondingEligibility))
	r.Get("/v1/global-params", registerHandler(handlers.V1.GetBabylonGlobalParams))
	r.Get("/v1/finality-providers", registerHandler(handlers.V1.GetFinalityProviders))
	r.Get("/v1/stats", registerHandler(handlers.V1.GetOverallStats))
	r.Get("/v1/stats/staker", registerHandler(handlers.V1.GetStakersStats))
	r.Get("/v1/staker/delegation/check", registerHandler(handlers.V1.CheckStakerDelegationExist))
	r.Get("/v1/delegation", registerHandler(handlers.V1.GetDelegationByTxHash))

	// Only register these routes if the asset has been configured
	// The endpoints are used to check ordinals within the UTXOs
	// Don't deprecate this endpoint
	if a.cfg.Assets != nil {
		r.Post("/v1/ordinals/verify-utxos", registerHandler(handlers.V1.VerifyUTXOs))
	}

	// Don't deprecate this endpoint
	r.Get("/v1/staker/pubkey-lookup", registerHandler(handlers.V1.GetPubKeys))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// V2 API
	// TODO: Implement the handlers for the V2 API
	r.Get("/v2/stats", registerHandler(handlers.V2.GetStats))
	r.Get("/v2/finality-providers", registerHandler(handlers.V2.GetFinalityProviders))
	r.Get("/v2/global-params", registerHandler(handlers.V2.GetGlobalParams))
	r.Get("/v2/staker/delegations", registerHandler(handlers.V2.GetStakerDelegations))
	r.Get("/v2/staker/stats", registerHandler(handlers.V2.GetStakerStats))
}
