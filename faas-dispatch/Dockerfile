FROM golang:1.15-alpine AS builder
WORKDIR /go/src/github.com/lolo-pop/faas-dispatch
COPY . .
RUN go build -o faas-dispatch

FROM alpine
WORKDIR /root/
COPY --from=builder /go/src/github.com/lolo-pop/faas-dispatch/faas-dispatch .
CMD ["./faas-dispatch"]