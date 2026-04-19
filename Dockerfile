FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
RUN adduser -D -u 1000 appuser

COPY --from=builder /api /usr/local/bin/api

USER appuser
EXPOSE 8080

ENTRYPOINT ["api"]
