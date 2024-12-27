# ------------------------------------------
# 1) Builder Stage
# ------------------------------------------
FROM golang:1.23-alpine3.19 AS builder
WORKDIR /app

# Copy module files first (to cache 'go mod download')
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the three binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app    ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o relay  ./cmd/relay
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o worker ./cmd/worker

# ------------------------------------------
# 2) Final Stage
# ------------------------------------------
FROM alpine:3.19
WORKDIR /app

# Copy compiled binaries from builder
COPY --from=builder /app/app    .
COPY --from=builder /app/relay  .
COPY --from=builder /app/worker .

# Default command
CMD ["./app"]
    