package main

import (
	"fmt"
	"log"

	"github.com/srbde/hive-anther/client"
)

func main() {
	// Initialize the client with Hive nodes
	nodes := []string{
		"https://api.hive.blog",
		"https://api.hivecosystem.dev",
	}
	api := client.NewClient(nodes, 30)

	fmt.Println("=== Anther Go Library - GetBlockRange Example ===")
	fmt.Println()

	// Get dynamic global properties to check current head block
	fmt.Println("Fetching current blockchain properties...")
	props, err := api.GetDynamicGlobalPropertiesStruct()
	if err != nil {
		log.Fatalf("Error fetching global properties: %v", err)
	}
	fmt.Printf("Current Head Block Number: %d\n", props.HeadBlockNumber)
	fmt.Println()

	// Fetch a range of 3 blocks starting 10 blocks behind the head
	startBlock := props.HeadBlockNumber - 10
	count := uint32(3)

	fmt.Printf("Fetching %d blocks starting from block %d...\n", count, startBlock)
	blocks, err := api.GetBlockRange(startBlock, count)
	if err != nil {
		log.Fatalf("Error fetching block range: %v", err)
	}

	fmt.Printf("✓ Successfully fetched %d blocks!\n", len(blocks))
	fmt.Println()

	// Print summary info for each block
	for i, block := range blocks {
		fmt.Printf("--- Block %d (Index %d in range) ---\n", startBlock+uint32(i), i)
		fmt.Printf("Block ID:   %s\n", block.BlockID)
		fmt.Printf("Witness:    %s\n", block.Witness)
		fmt.Printf("Timestamp:  %s\n", block.Timestamp)
		fmt.Printf("Tx Count:   %d\n", len(block.Transactions))
		fmt.Println()
	}

	fmt.Println("=== Example Complete ===")
}
