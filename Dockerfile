FROM alpine:3.21
RUN apk add --no-cache ca-certificates curl
ENV BRIDGE_SOURCE_CHAIN=zoo-evm BRIDGE_DEST_CHAIN=lux-c-chain BRIDGE_MPC_URL=http://zoo-mpc:8081 BRIDGE_BRAND_NAME=Zoo
COPY main.go /app/main.go
ENTRYPOINT ["echo", "Zoo Bridge — configure via K8s env vars"]
