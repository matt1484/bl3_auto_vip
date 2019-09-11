FROM golang

COPY . /go/src/github.com/matt1484/bl3_auto_vip
WORKDIR /go/src/github.com/matt1484/bl3_auto_vip

RUN go get github.com/thedevsaddam/gojsonq
RUN go get github.com/PuerkitoBio/goquery
RUN go get golang.org/x/crypto/ssh/terminal

CMD go run cmd/main.go