FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Copy go mod files (they're now at the root of the context)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the application from the cmd directory
RUN CGO_ENABLED=0 GOOS=linux go build -o pingspot ./cmd

FROM alpine:3.20
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/pingspot .

# Create upload directories
RUN mkdir -p /app/uploads/main /app/uploads/user

EXPOSE 8080

CMD ["./pingspot"]