# syntax=docker/dockerfile:1@sha256:ac85f380a63b13dfcefa89046420e1781752bab202122f8f50032edf31be0021
FROM golang:1.21-bullseye@sha256:54377c9f3aa6abd72185fdb95222496133792f4d3ea358fdbcb7b54b14db11f5 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build

FROM debian:bullseye-slim@sha256:8481471cacd29779e7b6fe64a891f4b6771d8c3c7e57bb410542a3db10d1897a
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/anilist-synchronizer ./
CMD ["./anilist-synchronizer"]
