FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY pkg/go.mod ./pkg/
COPY pkg/ ./pkg/

COPY attendance-service/go.mod attendance-service/go.sum ./attendance-service/
WORKDIR /build/attendance-service
RUN go mod download

COPY attendance-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /attendance-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /attendance-service .

EXPOSE 8085

CMD ["./attendance-service"]
