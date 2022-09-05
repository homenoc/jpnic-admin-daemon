## Build
FROM golang:1.18-bullseye AS build

WORKDIR /app
COPY . ./
RUN go mod download
WORKDIR /app/cmd/daemon
RUN go build -o /daemon
RUN ls /


## Deploy
FROM alpine:3

WORKDIR /
COPY --from=build /daemon /daemon
RUN ls
CMD ["/daemon","start"]