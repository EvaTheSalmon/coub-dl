# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /coub-dl .

FROM alpine:3.21
RUN apk add --no-cache ffmpeg ca-certificates \
    && adduser -D -H coub \
    && mkdir /data \
    && chown coub /data
COPY --from=build /coub-dl /usr/local/bin/coub-dl
USER coub
WORKDIR /data
ENTRYPOINT ["coub-dl"]
