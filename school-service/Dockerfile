FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY school-service/go.mod school-service/go.sum ./school-service/
WORKDIR /build/school-service
RUN go mod download

COPY school-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /school-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /school-service .

EXPOSE 8088

CMD ["./school-service"]
