FROM golang:1.24.13-alpine3.23 AS builder

ARG VERSION="HEAD"

RUN apk add --no-cache \
    make=4.4.1-r3 \
    git=2.52.0-r0 \
    build-base=0.5-r3 \
    linux-headers=6.16.12-r0 \
    musl-dev=1.2.5-r23 \
    pkgconf=2.5.1-r0 \
    alpine-sdk=1.1-r0 \
    libsodium-dev=1.0.20-r1 \
    libsodium-static=1.0.20-r1 \
    openssh=10.2_p1-r0

# Build
WORKDIR /go/src/github.com/babylonlabs-io/staking-api-service
COPY go.mod go.sum /go/src/github.com/babylonlabs-io/staking-api-service/
COPY ./ /go/src/github.com/babylonlabs-io/staking-api-service/

RUN LDFLAGS='-extldflags "-static" -v' \
    CGO_ENABLED=1 \
    CGO_LDFLAGS="$CGO_LDFLAGS -lstdc++ -lm -lsodium" \
    LINK_STATICALLY=true \
    BUILD_TAGS="muslc" \
    make build

# Final minimal image with binary only
FROM alpine:3.23 AS run

ARG VERSION="HEAD"

RUN addgroup --gid 1138 -S staking-api-service && adduser --uid 1138 -S staking-api-service -G staking-api-service
RUN apk add --no-cache \
    bash=5.3.3-r1 \
    curl=8.20.0-r0 \
    jq=1.8.1-r0

LABEL org.opencontainers.image.source="https://github.com/babylonlabs-io/staking-api-service:${VERSION}"

COPY --from=builder /go/src/github.com/babylonlabs-io/staking-api-service/build/staking-api-service /bin/staking-api-service
COPY --from=builder /go/src/github.com/babylonlabs-io/staking-api-service/contrib/images/staking-api-service/entrypoint.sh /bin/entrypoint.sh

WORKDIR /home/staking-api-service
RUN chown -R staking-api-service /home/staking-api-service
USER staking-api-service

ENTRYPOINT ["/bin/entrypoint.sh"]
CMD []
STOPSIGNAL SIGTERM