#!/usr/bin/env sh
set -euo pipefail
set -x

BINARY=${BINARY:-/bin/staking-api-service}
CONFIG=${CONFIG:-/home/staking-api-service/config.yml}
PARAMS=${PARAMS:-/home/staking-api-service/global-params.json}
FINALITY_PROVIDERS=${FINALITY_PROVIDERS:-/home/staking-api-service/finality-providers.json}

if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found."
	exit 1
fi

if ! [ -f "${CONFIG}" ]; then
	echo "The configuration file $(basename "${CONFIG}") cannot be found. Please add the configuration file to the shared folder. Use the CONFIG environment variable if the name of the configuration file is not 'config.yml'"
	exit 1
fi

if ! [ -f "${PARAMS}" ]; then
	echo "The global parameters file $(basename "${PARAMS}") cannot be found. Please add the global parameters file to the shared folder. Use the PARAMS environment variable if the name of the global parameters file is not 'global-params.json'"
	exit 1
fi

if ! [ -f "${FINALITY_PROVIDERS}" ]; then
	echo "The finality providers file $(basename "${FINALITY_PROVIDERS}") cannot be found. Please add the finality providers file to the shared folder. Use the FINALITY_PROVIDERS environment variable if the name of the finality providers file is not 'finality-providers.json'"
	exit 1
fi

$BINARY --config "$CONFIG" --params "$PARAMS" --finality-providers "$FINALITY_PROVIDERS" 2>&1
