# Stage 1: Build Stage
FROM golang:1.23-alpine AS builder

# Install necessary dependencies
RUN apk add --no-cache git build-base

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go application
RUN go build -o server .

# Stage 2: Final Stage
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Expose the application port
EXPOSE 8000

# Command to run the application
CMD ["./server"]