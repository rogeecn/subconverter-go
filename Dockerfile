FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o subconverter ./cmd/subconverter
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o subctl ./cmd/subctl
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o subworker ./cmd/subworker

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/subconverter .
COPY --from=builder /app/subctl .
COPY --from=builder /app/subworker .

# Copy configuration files
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/rules ./rules

# Create non-root user
RUN addgroup -g 1000 subconverter && \
    adduser -u 1000 -G subconverter -s /bin/sh -D subconverter

USER subconverter

EXPOSE 8080

CMD ["./subconverter"]