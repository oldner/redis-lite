FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o redis-lite ./cmd/server

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/redis-lite .
RUN mkdir data
EXPOSE 6379

ENV PORT=6379
ENV HOST=0.0.0.0
ENV AOF_PATH=./data/database.aof
ENV LOG_LEVEL=info

CMD ["./redis-lite"]
