## Build
FROM golang:1.19-bullseye AS build

WORKDIR /app
COPY . ./
RUN go mod download
WORKDIR /app/cmd/daemon
RUN CGO_ENABLED=0 go build -o /daemon

## Deploy
FROM alpine:3.16.2

WORKDIR /
RUN apk update&& \
    apk add --no-cache libc6-compat&& \
    rm -rf /var/cache/apk/*

COPY --from=build /daemon /daemon
CMD ["/daemon","start"]