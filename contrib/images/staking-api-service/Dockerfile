FROM golang:1.23.1-alpine AS builder

ARG VERSION="HEAD"

RUN apk add --no-cache  \
    make \
    git \
    build-base \
    linux-headers \
    libc-dev \
    pkgconfig \
    alpine-sdk \
    libsodium-dev \
    libsodium-static \
    openssh

# Build
WORKDIR /go/src/github.com/babylonlabs-io/staking-api-service
# Cache dependencies
COPY go.mod go.sum /go/src/github.com/babylonlabs-io/staking-api-service/

# Copy the rest of the files
COPY ./ /go/src/github.com/babylonlabs-io/staking-api-service/


RUN LDFLAGS='-extldflags "-static" -v' \
    CGO_ENABLED=1 \
    CGO_LDFLAGS="$CGO_LDFLAGS -lstdc++ -lm -lsodium" \
    LINK_STATICALLY=true \
    BUILD_TAGS="muslc" \
    make build

# Final minimal image with binary only
FROM alpine:3.16 as run

RUN addgroup --gid 1138 -S staking-api-service && adduser --uid 1138 -S staking-api-service -G staking-api-service
RUN apk add bash curl jq

# Label should match your github repo
LABEL org.opencontainers.image.source="https://github.com/babylonlabs-io/staking-api-service:${VERSION}"


# Copy over binaries from the build-env
COPY --from=builder /go/src/github.com/babylonlabs-io/staking-api-service/build/staking-api-service /bin/staking-api-service
COPY --from=builder /go/src/github.com/babylonlabs-io/staking-api-service/contrib/images/staking-api-service/entrypoint.sh /bin/entrypoint.sh

WORKDIR /home/staking-api-service
RUN chown -R staking-api-service /home/staking-api-service
USER staking-api-service

ENTRYPOINT ["/bin/entrypoint.sh"]
CMD []
STOPSIGNAL SIGTERM
