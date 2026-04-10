FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /zoo-bridge .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates curl
COPY --from=builder /zoo-bridge /usr/local/bin/zoo-bridge
ENV BRIDGE_LISTEN=:8080 \
    BRIDGE_UPSTREAM_URL=http://127.0.0.1:5000 \
    BRIDGE_BRAND_NAME=Zoo \
    BRIDGE_CORS_ORIGINS=https://bridge.zoo.ngo,https://zoo.ngo
EXPOSE 8080
RUN adduser -D -u 10001 zoo
USER zoo
ENTRYPOINT ["zoo-bridge"]
