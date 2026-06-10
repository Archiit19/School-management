FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY pkg/go.mod ./pkg/
COPY pkg/ ./pkg/

COPY finance-service/go.mod finance-service/go.sum ./finance-service/
WORKDIR /build/finance-service
RUN go mod download

COPY finance-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /finance-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /finance-service .

EXPOSE 8087

CMD ["./finance-service"]
