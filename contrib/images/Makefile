DOCKER = $(shell which docker)

.PHONY: staking-api-service staking-api-service-rmi

staking-api-service:
	$(DOCKER) build --tag babylonlabs-io/staking-api-service -f staking-api-service/Dockerfile \
				$(shell git rev-parse --show-toplevel)

staking-api-service-rmi:
	$(DOCKER) rmi babylonlabs-io/staking-api-service 2>/dev/null; true
