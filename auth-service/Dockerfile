FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY auth-service/go.mod auth-service/go.sum ./auth-service/
WORKDIR /build/auth-service
RUN go mod download

COPY auth-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /auth-service .

EXPOSE 8081

CMD ["./auth-service"]
