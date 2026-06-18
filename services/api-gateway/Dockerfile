# Dockerfile API Gateway ArticleSwap.

FROM golang:1.23-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY services/ ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-gateway ./cmd/api-gateway

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /out/api-gateway /app/api-gateway
EXPOSE 8080
CMD ["/app/api-gateway"]
