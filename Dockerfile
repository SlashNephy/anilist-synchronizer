# syntax=docker/dockerfile:1
FROM golang:1.20-bullseye@sha256:4c6e1c48e80c9562c8244750ffbfbd85be363772a37f76100eb48507f1910fa5 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build

FROM debian:bullseye-slim@sha256:7606bef5684b393434f06a50a3d1a09808fee5a0240d37da5d181b1b121e7637
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/anilist-synchronizer ./
CMD ["./anilist-synchronizer"]
