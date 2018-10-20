FROM golang:1.10-alpine as builder

ADD . /go/src/github.com/guilhermebr/demo
WORKDIR /go/src/github.com/guilhermebr/demo
RUN go build -o build/app .

FROM alpine:latest
COPY --from=builder /go/src/github.com/guilhermebr/demo/build/app /bin/
CMD ["./bin/app"]
