FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

ENV GOPRIVATE=github.com/Archiit19/School-management \
    GONOSUMDB=github.com/Archiit19/School-management

ARG GITHUB_TOKEN
RUN if [ -n "$GITHUB_TOKEN" ]; then \
      git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; \
    fi

WORKDIR /build

COPY pkg/go.mod pkg/go.sum ./pkg/
COPY academic-service/go.mod academic-service/go.sum ./academic-service/

WORKDIR /build/academic-service
RUN go mod download

COPY pkg/ /build/pkg/
COPY academic-service/ .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /academic-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /academic-service .

EXPOSE 8083

CMD ["./academic-service"]
