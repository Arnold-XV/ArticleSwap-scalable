# Dockerfile API Gateway ArticleSwap.

FROM golang:1.23-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY services/go.mod services/go.sum* ./
RUN go mod download
COPY services/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-gateway ./cmd/api-gateway

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /out/api-gateway /app/api-gateway
EXPOSE 8080
CMD ["/app/api-gateway"]
