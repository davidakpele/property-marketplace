FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

ENV GOTOOLCHAIN=auto

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/api \
    ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /build/api .
COPY --from=builder /build/configs ./configs
COPY --from=builder /build/migrations ./migrations

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

ENTRYPOINT ["./api"]