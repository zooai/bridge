FROM alpine:3.21
RUN apk add --no-cache ca-certificates curl
COPY zoo-bridge /usr/local/bin/zoo-bridge
ENV BRIDGE_LISTEN=:8080 \
    BRIDGE_UPSTREAM_URL=http://127.0.0.1:5000 \
    BRIDGE_BRAND_NAME=Zoo \
    BRIDGE_CORS_ORIGINS=https://bridge.zoo.ngo,https://zoo.ngo
EXPOSE 8080
RUN adduser -D -u 10001 zoo
USER zoo
ENTRYPOINT ["zoo-bridge"]
