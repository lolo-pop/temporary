FROM golang:1.15-alpine AS builder
WORKDIR /go/src/github.com/lolopop/test
COPY . .
RUN GOPROXY=https://goproxy.cn,direct go build -o test -mod=mod

FROM alpine
WORKDIR /root/
COPY --from=builder /go/src/github.com/lolopop/test/test .
CMD ["./test"]