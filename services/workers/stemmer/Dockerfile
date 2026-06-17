# Dockerfile Stemmer Worker ArticleSwap.

FROM golang:1.23-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY services/go.mod services/go.sum* ./
RUN go mod download
COPY services/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/stemmer-worker ./cmd/stemmer-worker

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/stemmer-worker /app/stemmer-worker
CMD ["/app/stemmer-worker"]
