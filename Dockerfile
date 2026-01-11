# syntax=docker/dockerfile:1

########################################
# 1) Builder Stage (Alpine + Go + CGO for sqlite3)
########################################
FROM golang:1.23.5-alpine AS builder

# Install C toolchain + sqlite headers for github.com/mattn/go-sqlite3
RUN apk add --no-cache build-base sqlite-dev

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source
COPY . .

RUN go build -ldflags="-s -w" -o forum main.go

# Clean up Go build cache
RUN rm -rf /root/.cache/go-build

########################################
# 2) Runtime Stage (Minimal Alpine)
########################################
FROM alpine:3.18

# Install only runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

# (Optional) Drop to a non-root user
RUN addgroup -S forum && adduser -S forum -G forum

USER forum
WORKDIR /app


# Copy your pre-built DB and set owner in one step
COPY --from=builder --chown=forum:forum /app/forum.db /app/forum.db

# Copy binary and assets from builder
COPY --from=builder /app/forum       /app/forum
COPY --from=builder /app/templates   /app/templates
COPY --from=builder /app/static      /app/static
# If you want runtime migrations/seeds:
COPY --from=builder /app/internal/database/migrations /app/internal/database/migrations
COPY --from=builder /app/internal/database/seeds      /app/internal/database/seeds


EXPOSE 8080

ENTRYPOINT ["./forum"]
