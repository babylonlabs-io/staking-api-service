# Logo/Icon Service - Project Scope Document

## Background

Babylon Labs currently fetches finality provider logos using the Keybase API. This approach has proven to be unreliable due to:
- Frequent API failures and downtime
- Unpredictable API changes
- No control over availability or performance
- Poor user experience with inconsistent logo loading

## Current Implementation

The Babylon Staking API Service (`babylon-staking-api-service`) currently:
- Fetches logos from Keybase API using identity strings from finality provider descriptions
- Caches logos in MongoDB with 1-day TTL
- Returns logo URLs in the `/v2/finality-providers` endpoint
- Uses asynchronous background fetching to avoid blocking API responses

**Key Files:**
- `internal/v2/service/finality_provider.go` - Logo fetching logic
- `internal/shared/integrations/keybase/client.go` - Keybase API integration
- `internal/v2/db/model/logo.go` - Logo data model

## Project Objective

Build a new centralized service to serve logos/icons from self-hosted AWS S3 storage with long-term browser caching. This service will:
1. Replace the unreliable Keybase API dependency
2. Provide consistent, fast logo delivery
3. Enable long-term browser caching for optimal performance

## Scope of Work

### 1. New Go Service Development (Prefer use existing side-car service)

Build a standalone Go service that:
- Provides REST API endpoints to fetch logos by unique identifier
- Serves images from AWS S3 storage
- Implements proper HTTP caching headers for long-term browser caching
- Returns appropriate status codes (200, 404, 500)
- Includes health check endpoint
- Is stateless and horizontally scalable
- Allow easy maintance of logos manually. i.e upload new logos, delete logos, etc via UI or API with admin privileges

**Key Requirements:**
- Written in Go (consistent with Babylon's tech stack)
- Simple HTTP API with endpoints like `GET /v1/logo/{identifier}`
- Cache-Control headers for 1-year browser caching
- ETag support for efficient cache validation

### 2. AWS Infrastructure Setup (Done by Babylon Labs devops team)

Set up and configure:
- **S3 Bucket**: For storing logo images in multiple sizes
- **CloudFront** (or equivalent CDN): For global distribution and caching

### 3. Documentation

Provide clear documentation including:
- README with setup and deployment instructions
- API documentation (OpenAPI/Swagger spec)
- Architecture overview diagram
- Migration guide with step-by-step instructions
- Configuration guide
- Troubleshooting guide

### 4. Deployment and DevOps (Done by Babylon Labs devops team)

Deliver deployment-ready artifacts:
- Docker container with multi-stage build
- Kubernetes manifests (Deployment, Service, Ingress, ConfigMap)
- CI/CD pipeline configuration (GitHub Actions or equivalent)
- Environment configuration templates

### 7. Monitoring and Observability

Implement basic monitoring:
- Prometheus metrics endpoint (request counts, latencies, errors)
- Structured JSON logging
- Health check endpoint for liveness/readiness probes

## Technical Requirements Summary

**Service Requirements:**
- Language: Go
- Framework: Standard Go HTTP server
- Storage: AWS S3

**API Requirements:**
- Simple REST API
- Long-term caching headers (1 year)
- Graceful error handling (404 for missing logos)
- HTTPS only

**Operational Requirements:**
- Containerized deployment
- Monitoring and logging