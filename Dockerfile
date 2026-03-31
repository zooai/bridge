# Zoo Bridge — upstream Lux Bridge with Zoo chain configuration
FROM ghcr.io/luxfi/bridge:latest
ENV BRIDGE_SOURCE_CHAIN=zoo-evm
ENV BRIDGE_DEST_CHAIN=lux-c-chain
ENV BRIDGE_MPC_URL=http://zoo-mpc:8081
ENV BRIDGE_BRAND_NAME=Zoo
