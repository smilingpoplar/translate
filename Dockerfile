FROM golang:1.21-alpine3.18 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd/translate


FROM alpine:3.18
COPY --from=builder /app/main /main

ENTRYPOINT ["/main"]
