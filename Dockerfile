# Build stage
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Pre-download dependencies for better caching
COPY go.mod go.sum ./
RUN go mod download

# Build the source code
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o slack-explorer-mcp .

# Runtime stage
FROM alpine:latest

# MCP Registry metadata
LABEL io.modelcontextprotocol.server.name="io.github.shibayu36/slack-explorer-mcp"

# Add CA certificates for HTTPS communication
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the built binary
COPY --from=builder /app/slack-explorer-mcp .

# Set the entrypoint
ENTRYPOINT ["./slack-explorer-mcp"]
