# syntax=docker/dockerfile:1
FROM golang:latest AS build-env
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.mod /src/
RUN go mod download
COPY . .
RUN go build -a -o gobuster -trimpath

FROM alpine:latest

ARG UID=1000
ARG GID=1000

RUN apk add --no-cache ca-certificates \
    && rm -rf /var/cache/*

RUN mkdir -p /app \
    && addgroup -g ${GID} gobuster \
    && adduser -D gobuster -u ${UID} -G gobuster \
    && chown -R gobuster:gobuster /app

USER ${UID}:${GID}
WORKDIR /app

COPY --from=build-env /src/gobuster .

ENTRYPOINT [ "./gobuster" ]
