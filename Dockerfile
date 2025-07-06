FROM golang:1.24-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run ./cmd/schema-fetcher --url https://raw.githubusercontent.com/first-debug/lk-graphql-schemas/master/schemas/user-provider/schema.graphql --output api/graphql/schema.graphql

RUN go generate ./...

RUN CGO_ENABLE=0 go build -ldflags="-w -s" -o /lk-auth ./cmd/main.go

FROM alpine:latest

COPY --from=builder /lk-auth /lk-auth

COPY config/config_local.yml /config/config_local.yml
COPY .env /.env

WORKDIR /

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 80

CMD [ "/lk-auth" ]
