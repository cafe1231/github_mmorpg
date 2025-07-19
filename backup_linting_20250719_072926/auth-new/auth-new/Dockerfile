# Dockerfile pour le service auth-new
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/main ./main
COPY --from=builder /app/internal ./internal
COPY --from=builder /app/go.mod ./go.mod
COPY --from=builder /app/go.sum ./go.sum
EXPOSE 8081
CMD ["./main"] 