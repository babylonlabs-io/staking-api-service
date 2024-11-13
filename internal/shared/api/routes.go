package api

import (
	_ "github.com/babylonlabs-io/staking-api-service/docs"
	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (a *Server) SetupRoutes(r *chi.Mux) {
	handlers := a.handlers
	// Extend on the healthcheck endpoint here
	r.Get("/healthcheck", registerHandler(handlers.SharedHandler.HealthCheck))

	r.Get("/v1/staker/delegations", registerHandler(handlers.V1Handler.GetStakerDelegations))
	r.Post("/v1/unbonding", registerHandler(handlers.V1Handler.UnbondDelegation))
	r.Get("/v1/unbonding/eligibility", registerHandler(handlers.V1Handler.GetUnbondingEligibility))
	r.Get("/v1/global-params", registerHandler(handlers.V1Handler.GetBabylonGlobalParams))
	r.Get("/v1/finality-providers", registerHandler(handlers.V1Handler.GetFinalityProviders))
	r.Get("/v1/stats", registerHandler(handlers.V1Handler.GetOverallStats))
	r.Get("/v1/stats/staker", registerHandler(handlers.V1Handler.GetStakersStats))
	r.Get("/v1/staker/delegation/check", registerHandler(handlers.V1Handler.CheckStakerDelegationExist))
	r.Get("/v1/delegation", registerHandler(handlers.V1Handler.GetDelegationByTxHash))

	// Only register these routes if the asset has been configured
	// The endpoints are used to check ordinals within the UTXOs
	// Don't deprecate this endpoint
	if a.cfg.Assets != nil {
		r.Post("/v1/ordinals/verify-utxos", registerHandler(handlers.SharedHandler.VerifyUTXOs))
	}

	// Don't deprecate this endpoint
	r.Get("/v1/staker/pubkey-lookup", registerHandler(handlers.V1Handler.GetPubKeys))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// V2 API
	r.Get("/v2/stats", registerHandler(handlers.V2Handler.GetStats))
	r.Get("/v2/finality-providers", registerHandler(handlers.V2Handler.GetFinalityProviders))
	r.Get("/v2/params", registerHandler(handlers.V2Handler.GetParams))
	r.Get("/v2/delegation", registerHandler(handlers.V2Handler.GetDelegation))
	r.Get("/v2/delegations", registerHandler(handlers.V2Handler.GetDelegations))
	r.Get("/v2/staker/stats", registerHandler(handlers.V2Handler.GetStakerStats))
}
