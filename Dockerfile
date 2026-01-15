FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pingspot ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/pingspot .

RUN mkdir -p /app/uploads/main /app/uploads/user

EXPOSE 8080

CMD ["./pingspot"]