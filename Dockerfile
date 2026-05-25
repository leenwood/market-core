FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/server ./cmd/server && \
    go build -o bin/migrate ./cmd/migrate

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/bin/server ./server
COPY --from=builder /app/bin/migrate ./migrate
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["sh", "-c", "./migrate -dir=migrations && ./server"]
