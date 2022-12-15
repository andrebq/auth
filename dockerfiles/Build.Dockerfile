FROM golang:1

WORKDIR /app/auth
COPY . .
RUN go build -o /usr/local/bin/auth ./cmd/auth
