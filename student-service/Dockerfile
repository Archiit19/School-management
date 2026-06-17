FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY student-service/go.mod student-service/go.sum ./student-service/
WORKDIR /build/student-service
RUN go mod download

COPY student-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /student-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /student-service .

EXPOSE 8084

CMD ["./student-service"]
