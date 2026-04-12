FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run scripts/keys.go

RUN CGO_ENABLED=0 GOOS=linux go build -o pingspot ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

RUN mkdir -p /app/uploads/main /app/uploads/user

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/pingspot .
COPY --from=builder /app/keys ./keys

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./pingspot"]