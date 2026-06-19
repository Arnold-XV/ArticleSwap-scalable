# Dockerfile Wordcloud Worker ArticleSwap.

FROM golang:1.23-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY services/ ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/wordcloud-worker ./cmd/wordcloud-worker

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/wordcloud-worker /app/wordcloud-worker
CMD ["/app/wordcloud-worker"]
