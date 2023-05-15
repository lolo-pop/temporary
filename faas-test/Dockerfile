FROM golang:1.15-alpine AS builder
WORKDIR /go/src/github.com/smvfal/faas-test
COPY . .
RUN GOPROXY=https://goproxy.cn,direct go build -o faas-test -mod=mod

FROM alpine
WORKDIR /root/
COPY --from=builder /go/src/github.com/smvfal/faas-test/faas-test .
CMD ["./faas-test"]