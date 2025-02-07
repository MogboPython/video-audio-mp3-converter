# Use a multi-stage build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

# Final stage
FROM alpine:latest

# Install ffmpeg
RUN apk add --no-cache ffmpeg

# Copy binary from builder
WORKDIR /app
COPY --from=builder /app/main .
COPY .env.development .
COPY credentials.dev.json .

# Set environment variables
ENV PORT=8080
ENV GO_ENV=development

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"] 