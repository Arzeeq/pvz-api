# Stage builder
FROM golang:1.23.8-alpine as builder
WORKDIR /avito
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o pvz-api ./cmd/pvz-api \
    && go clean -cache -modcache

# Stage base
FROM alpine:latest as base
WORKDIR /root
COPY --from=builder /avito/pvz-api .
COPY --from=builder /avito/internal/storage/pg/migrations ./migrations
COPY --from=builder /avito/configs ./configs
RUN apk --no-cache add ca-certificates
EXPOSE 8080
CMD ["./pvz-api"]