// Copyright (C) 2026, Zoo Labs Foundation. All rights reserved.
// Zoo Bridge — Cross-chain bridge for Zoo EVM <> Lux C-Chain <> external chains.
// Uses Zoo MPC for transaction signing.
package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version") {
		fmt.Printf("zoo-bridge %s\n", version)
		os.Exit(0)
	}

	fmt.Println("Zoo Bridge — use ghcr.io/luxfi/bridge with Zoo configuration")
	fmt.Println("Environment variables:")
	fmt.Println("  BRIDGE_SOURCE_CHAIN=zoo-evm")
	fmt.Println("  BRIDGE_DEST_CHAIN=lux-c-chain")
	fmt.Println("  BRIDGE_MPC_URL=http://zoo-mpc:8081")
	fmt.Println("  BRIDGE_BRAND_NAME=Zoo")
}
