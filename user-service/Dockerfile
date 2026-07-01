FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

ENV GOPRIVATE=github.com/Archiit19/School-management \
    GONOSUMDB=github.com/Archiit19/School-management

ARG GITHUB_TOKEN
RUN if [ -n "$GITHUB_TOKEN" ]; then \
      git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; \
    fi

WORKDIR /build

# Local pkg module (includes tracer); go.mod uses replace => ../pkg
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY user-service/go.mod user-service/go.sum ./user-service/

WORKDIR /build/user-service
RUN go mod download

COPY pkg/ /build/pkg/
COPY user-service/ .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /user-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /user-service .

EXPOSE 8082

CMD ["./user-service"]
