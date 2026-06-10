FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY pkg/go.mod ./pkg/
COPY pkg/ ./pkg/

COPY academic-service/go.mod academic-service/go.sum ./academic-service/
WORKDIR /build/academic-service
RUN go mod download

COPY academic-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /academic-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /academic-service .

EXPOSE 8083

CMD ["./academic-service"]
