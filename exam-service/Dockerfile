FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY pkg/go.mod ./pkg/
COPY pkg/ ./pkg/

COPY exam-service/go.mod exam-service/go.sum ./exam-service/
WORKDIR /build/exam-service
RUN go mod download

COPY exam-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /exam-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /exam-service .

EXPOSE 8086

CMD ["./exam-service"]
