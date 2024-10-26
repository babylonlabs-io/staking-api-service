package api

import (
	_ "github.com/babylonlabs-io/staking-api-service/docs"
	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (a *Server) SetupRoutes(r *chi.Mux) {
	handlers := a.handlers
	r.Get("/healthcheck", registerHandler(handlers.HealthCheck))

	if a.cfg.TermsAcceptanceLogging.Enabled {
		r.Post("/log_terms_acceptance", registerHandler(handlers.LogTermsAcceptance))
	}

	r.Get("/v1/staker/delegations", registerHandler(handlers.GetStakerDelegations))
	r.Post("/v1/unbonding", registerHandler(handlers.UnbondDelegation))
	r.Get("/v1/unbonding/eligibility", registerHandler(handlers.GetUnbondingEligibility))
	r.Get("/v1/global-params", registerHandler(handlers.GetBabylonGlobalParams))
	r.Get("/v1/finality-providers", registerHandler(handlers.GetFinalityProviders))
	r.Get("/v1/stats", registerHandler(handlers.GetOverallStats))
	r.Get("/v1/stats/staker", registerHandler(handlers.GetStakersStats))
	r.Get("/v1/staker/delegation/check", registerHandler(handlers.CheckStakerDelegationExist))
	r.Get("/v1/delegation", registerHandler(handlers.GetDelegationByTxHash))

	// Only register these routes if the asset has been configured
	// The endpoints are used to check ordinals within the UTXOs
	if a.cfg.Assets != nil {
		r.Post("/v1/ordinals/verify-utxos", registerHandler(handlers.VerifyUTXOs))
	}

	r.Get("/v1/staker/pubkey-lookup", registerHandler(handlers.GetPubKeys))

	r.Get("/swagger/*", httpSwagger.WrapHandler)
}
