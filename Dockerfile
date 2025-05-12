FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main.exe main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main.exe .

RUN apk add --no-cache tcpdump
RUN chmod +x /app/main.exe

ENTRYPOINT ["/app/main.exe"]