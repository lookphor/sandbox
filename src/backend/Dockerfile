ARG VERSION="0.0.1"

# Стадия builder
FROM golang:1.22-alpine AS builder

ARG VERSION                       # ← переобъявлен после FROM

RUN apk add --no-cache git

WORKDIR /app

# Кеш зависимостей отдельным слоем
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-w -s" \
    -o "./build/backend-${VERSION}/app" \
    ./cmd/api

# runtime
FROM golang:1.22-alpine AS runtime

ARG VERSION                       # ← переобъявлен снова

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder "/app/build/backend-${VERSION}/app" /app/app

USER app
EXPOSE 8081
ENTRYPOINT ["/app/app"]