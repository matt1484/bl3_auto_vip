FROM golang:1.12-alpine

COPY . /go/src/github.com/matt1484/bl3_auto_vip
WORKDIR /go/src/github.com/matt1484/bl3_auto_vip

ENV GO111MODULE=on

RUN apk add git
RUN go mod download && go mod verify

CMD go run cmd/main.go