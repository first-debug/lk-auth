FROM golang:1.24-alpine AS builder

WORKDIR /build-dir

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLE=0 go build -ldflags="-w -s" -o ./lk-auth ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /build-dir/lk-auth ./start

# -v ($pwd)/config:/app/config
# --env-file .env

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 80

CMD [ "/app/start" ]
