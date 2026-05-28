FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o subscription-aggregator-api ./cmd/app

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/subscription-aggregator-api .

EXPOSE 8080

CMD ["./subscription-aggregator-api"]