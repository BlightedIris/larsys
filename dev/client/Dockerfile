# Build stage
FROM golang:1.24-bookworm AS builder

WORKDIR /build

# Copy go.mod (and go.sum if it exists)
COPY go/go.mod ./
COPY go/go.sum* ./

# Copy source
COPY go/ ./

# Build binary for Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o larsys-client ./client

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/larsys-client /usr/bin/larsys

RUN mkdir -p /var/log/larsys && chmod 755 /var/log/larsys

COPY dev/client/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 5454

CMD ["/usr/local/bin/entrypoint.sh"]