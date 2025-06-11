# Multi-stage build untuk optimasi ukuran image
FROM golang:1.19-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (diperlukan untuk go mod download)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dan install dependencies
RUN go mod download && go mod tidy

# Copy source code
COPY . .

# Build aplikasi
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage kedua - runtime image
FROM alpine:latest

# Install ca-certificates untuk HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app directory dan user
RUN mkdir -p /app/uploads/images /app/uploads/product-images /app/uploads/store-avatar /app/uploads/store-banner && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary dari builder stage
COPY --from=builder /app/main .

# Set permissions untuk uploads directory
RUN chown -R appuser:appgroup /app && \
    chmod -R 755 /app/uploads

# Switch ke non-root user
USER appuser

# Expose port
EXPOSE 1815

# Command untuk menjalankan aplikasi
CMD ["./main"]