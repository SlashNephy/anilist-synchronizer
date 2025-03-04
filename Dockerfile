# syntax=docker/dockerfile:1@sha256:93bfd3b68c109427185cd78b4779fc82b484b0b7618e36d0f104d4d801e66d25
FROM golang:1.24-bullseye@sha256:aa106963247f64275bd459b6b713978f1633160da53f58115922964ab0b9eae7 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build

FROM debian:bullseye-slim@sha256:33b7c2e071c29e618182ec872c471f39d2dde3d8904d95f5b7a61acf3a592e7b
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/anilist-synchronizer ./
CMD ["./anilist-synchronizer"]
