// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "email": "contact@babylonlabs.io"
        },
        "license": {
            "name": "API Access License",
            "url": "https://docs.babylonlabs.io/assets/files/api-access-license.pdf"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/healthcheck": {
            "get": {
                "description": "Health check the service, including ping database connection",
                "produces": [
                    "application/json"
                ],
                "summary": "Health check endpoint",
                "responses": {
                    "200": {
                        "description": "Server is up and running",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/delegation": {
            "get": {
                "description": "Retrieves a delegation by a given transaction hash",
                "produces": [
                    "application/json"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Staking transaction hash in hex format",
                        "name": "staking_tx_hash_hex",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Delegation",
                        "schema": {
                            "$ref": "#/definitions/handlers.PublicResponse-services_DelegationPublic"
                        }
                    },
                    "400": {
                        "description": "Error: Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_babylonlabs-io_staking-api-service_internal_types.Error"
                        }
                    }
                }
            }
        },
        "/v1/finality-providers": {
            "get": {
                "description": "Fetches details of all active finality providers sorted by their active total value locked (ActiveTvl) in descending order.",
                "produces": [
                    "application/json"
                ],
                "summary": "Get Active Finality Providers",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Public key of the finality provider to fetch",
                        "name": "fp_btc_pk",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Pagination key to fetch the next page of finality providers",
                        "name": "pagination_key",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "A list of finality providers sorted by ActiveTvl in descending order",
                        "schema": {
                            "$ref": "#/definitions/handlers.PublicResponse-array_services_FpDetailsPublic"
                        }
                    }
                }
            }
        },
        "/v1/global-params": {
            "get": {
                "description": "Retrieves the global parameters for Babylon, including finality provider details.",
                "produces": [
                    "application/json"
                ],
                "summary": "Get Babylon global parameters",
                "responses": {
                    "200": {
                        "description": "Global parameters",
                        "schema": {
                            "$ref": "#/definitions/handlers.PublicResponse-services_GlobalParamsPublic"
                        }
                    }
                }
            }
        },
        "/v1/staker/delegation/check": {
            "get": {
                "description": "Check if a staker has an active delegation by the staker BTC address (Taproot or Native Segwit)\nOptionally, you can provide a timeframe to check if the delegation is active within the provided timeframe\nThe available timeframe is \"today\" which checks after UTC 12AM of the current day",
                "produces": [
                    "application/json"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Staker BTC address in Taproot/Native Segwit format",
                        "name": "address",
                        "in": "query",
                        "required": true
                    },
                    {
                        "enum": [
                            "today"
                        ],
                        "type": "string",
                        "description": "Check if the delegation is active within the provided timeframe",
                        "name": "timeframe",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Delegation check result",
                        "schema": {
                            "$ref": "#/definitions/handlers.DelegationCheckPublicResponse"
                        }
                    },
                    "400": {
                        "description": "Error: Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_babylonlabs-io_staking-api-service_internal_types.Error"
                        }
                    }
                }
            }
        },
        "/v1/staker/delegations": {
            "get": {
                "description": "Retrieves delegations for a given staker",
                "produces": [
                    "application/json"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Staker BTC Public Key",
                        "name": "staker_btc_pk",
                        "in": "query",
                        "required": true
                    },
                    {
                        "enum": [
                            "active",
                            "unbonding_requested",
                            "unbonding",
                            "unbonded",
                            "withdrawn"
                        ],
                        "type": "string",
                        "description": "Filter by state",
                        "name": "state",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Pagination key to fetch the next page of delegations",
                        "name": "pagination_key",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "List of delegations and pagination token",
                        "schema": {
                            "$ref": "#/definitions/handlers.PublicResponse-array_services_DelegationPublic"
                        }
                    },
                    "400": {
                        "description": "Error: Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_babylonlabs-io_staking-api-service_internal_types.Error"
                        }
                    }
                }
            }
        },
        "/v1/staker/pubkey-lookup": {
            "get": {
                "description": "Retrieves public keys for the given BTC addresses. This endpoint",
                "produces": [
                    "application/json"
                ],
                "parameters": [
                    {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "collectionFormat": "multi",
                        "description": "List of BTC addresses to look up (up to 10), currently only supports Taproot and Native Segwit addresses",
                        "name": "address",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "A map of BTC addresses to their corresponding public keys (only addresses with delegations are returned)",
                        "schema": {
                            "$ref": "#/definitions/handlers.Result"
                        }
                    },
                    "400": {
                        "description": "Bad Request: Invalid input parameters",
                        "schema": {
                            "$ref": "#/definitions/github_com_babylonlabs-io_staking-api-service_internal_types.Error"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/github_com_babylonlabs-io_staking-api-service_internal_types.Error"
                        }
                    }
                }
            }
        },
        "/v1/stats": {
            "get": {
                "description": "Fetches overall stats for babylon staking including tvl, total delegations, active tvl, active delegations and total stakers.",
                "produces": [
                    "application/json"
                ],
                "summary": "Get Overall Stats",
                "responses": {
                    "200": {
                        "description": "Overall stats for babylon staking",
                        "schema": {
                            "$ref": "#/definitions/handlers.PublicResponse-services_OverallStatsPublic"
                        }
                    }
                }
            }
        },
        "/v1/stats/staker": {
            "get": {
                "description": "Fetches staker stats for babylon staking including tvl, total delegations, active tvl and active delegations.\nIf staker_btc_pk query parameter is provided, it will return stats for the specific staker.\nOtherwise, it will return the top stakers ranked by active tvl.",
                "produces": [
                    "application/json"
                ],
                "summary": "Get Staker Stats",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Public key of the staker to fetch",
                        "name": "staker_btc_pk",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Pagination key to fetch the next page of top stakers",
                        "name": "pagination_key",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "List of top stakers by active tvl",
                        "schema": {
                            "$ref": "#/definitions/handlers.PublicResponse-array_services_StakerStatsPublic"
                        }
                    },
                    "400": {
                        "description": "Error: Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_babylonlabs-io_staking-api-service_internal_types.Error"
                        }
                    }
                }
            }
        },
        "/v1/unbonding": {
            "post": {
                "description": "Unbonds a delegation by processing the provided transaction details. This is an async operation.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Unbond delegation",
                "parameters": [
                    {
                        "description": "Unbonding Request Payload",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.UnbondDelegationRequestPayload"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Request accepted and will be processed asynchronously"
                    },
                    "400": {
                        "description": "Invalid request payload",
                        "schema": {
                            "$ref": "#/definitions/github_com_babylonlabs-io_staking-api-service_internal_types.Error"
                        }
                    }
                }
            }
        },
        "/v1/unbonding/eligibility": {
            "get": {
                "description": "Checks if a delegation identified by its staking transaction hash is eligible for unbonding.",
                "produces": [
                    "application/json"
                ],
                "summary": "Check unbonding eligibility",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Staking Transaction Hash Hex",
                        "name": "staking_tx_hash_hex",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "The delegation is eligible for unbonding"
                    },
                    "400": {
                        "description": "Missing or invalid 'staking_tx_hash_hex' query parameter",
                        "schema": {
                            "$ref": "#/definitions/github_com_babylonlabs-io_staking-api-service_internal_types.Error"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "github_com_babylonlabs-io_staking-api-service_internal_types.Error": {
            "type": "object",
            "properties": {
                "err": {},
                "errorCode": {
                    "$ref": "#/definitions/types.ErrorCode"
                },
                "statusCode": {
                    "type": "integer"
                }
            }
        },
        "handlers.DelegationCheckPublicResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {
                    "type": "boolean"
                }
            }
        },
        "handlers.PublicResponse-array_services_DelegationPublic": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/services.DelegationPublic"
                    }
                },
                "pagination": {
                    "$ref": "#/definitions/handlers.paginationResponse"
                }
            }
        },
        "handlers.PublicResponse-array_services_FpDetailsPublic": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/services.FpDetailsPublic"
                    }
                },
                "pagination": {
                    "$ref": "#/definitions/handlers.paginationResponse"
                }
            }
        },
        "handlers.PublicResponse-array_services_StakerStatsPublic": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/services.StakerStatsPublic"
                    }
                },
                "pagination": {
                    "$ref": "#/definitions/handlers.paginationResponse"
                }
            }
        },
        "handlers.PublicResponse-services_DelegationPublic": {
            "type": "object",
            "properties": {
                "data": {
                    "$ref": "#/definitions/services.DelegationPublic"
                },
                "pagination": {
                    "$ref": "#/definitions/handlers.paginationResponse"
                }
            }
        },
        "handlers.PublicResponse-services_GlobalParamsPublic": {
            "type": "object",
            "properties": {
                "data": {
                    "$ref": "#/definitions/services.GlobalParamsPublic"
                },
                "pagination": {
                    "$ref": "#/definitions/handlers.paginationResponse"
                }
            }
        },
        "handlers.PublicResponse-services_OverallStatsPublic": {
            "type": "object",
            "properties": {
                "data": {
                    "$ref": "#/definitions/services.OverallStatsPublic"
                },
                "pagination": {
                    "$ref": "#/definitions/handlers.paginationResponse"
                }
            }
        },
        "handlers.Result": {
            "type": "object",
            "properties": {
                "data": {},
                "status": {
                    "type": "integer"
                }
            }
        },
        "handlers.UnbondDelegationRequestPayload": {
            "type": "object",
            "properties": {
                "staker_signed_signature_hex": {
                    "type": "string"
                },
                "staking_tx_hash_hex": {
                    "type": "string"
                },
                "unbonding_tx_hash_hex": {
                    "type": "string"
                },
                "unbonding_tx_hex": {
                    "type": "string"
                }
            }
        },
        "handlers.paginationResponse": {
            "type": "object",
            "properties": {
                "next_key": {
                    "type": "string"
                }
            }
        },
        "services.DelegationPublic": {
            "type": "object",
            "properties": {
                "finality_provider_pk_hex": {
                    "type": "string"
                },
                "is_overflow": {
                    "type": "boolean"
                },
                "staker_pk_hex": {
                    "type": "string"
                },
                "staking_tx": {
                    "$ref": "#/definitions/services.TransactionPublic"
                },
                "staking_tx_hash_hex": {
                    "type": "string"
                },
                "staking_value": {
                    "type": "integer"
                },
                "state": {
                    "type": "string"
                },
                "unbonding_tx": {
                    "$ref": "#/definitions/services.TransactionPublic"
                }
            }
        },
        "services.FpDescriptionPublic": {
            "type": "object",
            "properties": {
                "details": {
                    "type": "string"
                },
                "identity": {
                    "type": "string"
                },
                "moniker": {
                    "type": "string"
                },
                "security_contact": {
                    "type": "string"
                },
                "website": {
                    "type": "string"
                }
            }
        },
        "services.FpDetailsPublic": {
            "type": "object",
            "properties": {
                "active_delegations": {
                    "type": "integer"
                },
                "active_tvl": {
                    "type": "integer"
                },
                "btc_pk": {
                    "type": "string"
                },
                "commission": {
                    "type": "string"
                },
                "description": {
                    "$ref": "#/definitions/services.FpDescriptionPublic"
                },
                "total_delegations": {
                    "type": "integer"
                },
                "total_tvl": {
                    "type": "integer"
                }
            }
        },
        "services.GlobalParamsPublic": {
            "type": "object",
            "properties": {
                "versions": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/services.VersionedGlobalParamsPublic"
                    }
                }
            }
        },
        "services.OverallStatsPublic": {
            "type": "object",
            "properties": {
                "active_delegations": {
                    "type": "integer"
                },
                "active_tvl": {
                    "type": "integer"
                },
                "pending_tvl": {
                    "type": "integer"
                },
                "total_delegations": {
                    "type": "integer"
                },
                "total_stakers": {
                    "type": "integer"
                },
                "total_tvl": {
                    "type": "integer"
                },
                "unconfirmed_tvl": {
                    "type": "integer"
                }
            }
        },
        "services.StakerStatsPublic": {
            "type": "object",
            "properties": {
                "active_delegations": {
                    "type": "integer"
                },
                "active_tvl": {
                    "type": "integer"
                },
                "staker_pk_hex": {
                    "type": "string"
                },
                "total_delegations": {
                    "type": "integer"
                },
                "total_tvl": {
                    "type": "integer"
                }
            }
        },
        "services.TransactionPublic": {
            "type": "object",
            "properties": {
                "output_index": {
                    "type": "integer"
                },
                "start_height": {
                    "type": "integer"
                },
                "start_timestamp": {
                    "type": "string"
                },
                "timelock": {
                    "type": "integer"
                },
                "tx_hex": {
                    "type": "string"
                }
            }
        },
        "services.VersionedGlobalParamsPublic": {
            "type": "object",
            "properties": {
                "activation_height": {
                    "type": "integer"
                },
                "cap_height": {
                    "type": "integer"
                },
                "confirmation_depth": {
                    "type": "integer"
                },
                "covenant_pks": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "covenant_quorum": {
                    "type": "integer"
                },
                "max_staking_amount": {
                    "type": "integer"
                },
                "max_staking_time": {
                    "type": "integer"
                },
                "min_staking_amount": {
                    "type": "integer"
                },
                "min_staking_time": {
                    "type": "integer"
                },
                "staking_cap": {
                    "type": "integer"
                },
                "tag": {
                    "type": "string"
                },
                "unbonding_fee": {
                    "type": "integer"
                },
                "unbonding_time": {
                    "type": "integer"
                },
                "version": {
                    "type": "integer"
                }
            }
        },
        "types.ErrorCode": {
            "type": "string",
            "enum": [
                "INTERNAL_SERVICE_ERROR",
                "VALIDATION_ERROR",
                "NOT_FOUND",
                "BAD_REQUEST",
                "FORBIDDEN",
                "UNPROCESSABLE_ENTITY",
                "REQUEST_TIMEOUT"
            ],
            "x-enum-varnames": [
                "InternalServiceError",
                "ValidationError",
                "NotFound",
                "BadRequest",
                "Forbidden",
                "UnprocessableEntity",
                "RequestTimeout"
            ]
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Babylon Staking API",
	Description:      "The Babylon Staking API offers information about the state of the Phase-1 BTC Staking system.\nYour access and use is governed by the API Access License linked to below.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
