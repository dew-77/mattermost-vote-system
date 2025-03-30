FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev openssl-dev git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -v -o bot ./cmd/bot

FROM alpine:latest

RUN apk --no-cache add ca-certificates openssl

WORKDIR /root/

COPY --from=builder /app/bot .
COPY config.yaml .

CMD ["./bot"]