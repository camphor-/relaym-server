FROM golang:1.18.3 AS build-env

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /go/src/github.com/camphor-/relaym-server

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build .

FROM alpine:3.11.6

RUN apk add --no-cache bash ca-certificates curl

COPY ./docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod a+x /docker-entrypoint.sh

COPY ./mysql /mysql

COPY --from=build-env /go/src/github.com/camphor-/relaym-server/relaym-server /relaym-server
RUN chmod a+x /relaym-server

EXPOSE 8080
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/relaym-server"]
