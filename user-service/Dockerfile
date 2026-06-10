FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY pkg/go.mod ./pkg/
COPY pkg/ ./pkg/

COPY user-service/go.mod user-service/go.sum ./user-service/
WORKDIR /build/user-service
RUN go mod download

COPY user-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /user-service .

EXPOSE 8082

CMD ["./user-service"]
