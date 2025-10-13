## Project Overview

This is a backend service that provides API for data that is fetched and indexed by another service called `babylon-staking-indexer`.
It's part of the backend services for cosmos network called Babylon ("bbn" for short)

## Overview

- This project is written in Go
- Documentation (including swagger) is located in `docs/` directory (for example state transition is described there)
- Configuration files are located in `config/` directory (`config-local.yml` is used for `make run-local`)

## Dependencies

- MongoDB as database
- babylon node
- rabbitmq as message bus
- Different 3rd party APIs that are used in few places (like price fetching etc.)

## Development workflow

- Every time you're done with code changes run `make lint` to verify that linter doesn't complain about errors
- If there are formatting errors you must fix them by calling `make format`
- Do not modify mocks manually - always call `make generate` (it will generate code for interfaces using mockery). Mocks located in `tests/mocks/`
- If you want to run it locally just call `make run-local` (do not use make commands for docker)
- If you modify API request/response please update comment above corresponding handler it's used for autogenerating swagger docs (`make build-swagger` command).
Once you update API comment run `make build-swagger` so the doc is up to date.

## Testing

If you need to test that everything is working you should call one of these commands:
- `make test` for unit-tests
- `make test-integration` for tests related to database (note that it will set up a docker container with mongo so it's a bit time consuming)
- `make test-e2e` this is an end-to-end test for the API that sets up all dependencies, sends a request to the real server and compares it with the expected response