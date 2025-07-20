# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum from current directory
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy Go source code from backend folder
COPY backend/ ./backend/

# Set the working directory to backend 
WORKDIR /app/backend

# Build the app binary
RUN go build -o /app/server main.go

# Stage 2: Run
FROM alpine:latest

WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
