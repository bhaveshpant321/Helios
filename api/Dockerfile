# Step 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy dependency files and download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Step 2: Create the production image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./main"]
