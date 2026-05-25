package main

import (
	"fmt"
	"log"

	"github.com/thecrazygm/anther/client"
)

func main() {
	// Initialize the client with a public Hive node
	nodes := []string{"https://api.hive.blog"}
	api := client.NewClient(nodes, 30)

	fmt.Println("=== Anther Go Library - Advanced API Examples ===")
	fmt.Println()

	// 1. Database Parameters & Config
	fmt.Println("🌐 [1] Fetching Database Parameters & Configuration...")
	config, err := api.GetConfig()
	if err != nil {
		log.Fatalf("Error fetching config: %v", err)
	}
	// Print a few key configuration values
	fmt.Printf("✓ Hive Chain ID:           %v\n", config["HIVE_CHAIN_ID"])
	fmt.Printf("  Max Block Size:          %v bytes\n", config["HIVE_MAX_BLOCK_SIZE"])
	fmt.Printf("  Vesting Withdraw Rate:   %v\n", config["HIVE_VESTING_WITHDRAW_INTERVALS"])
	fmt.Println()

	props, err := api.GetChainProperties()
	if err != nil {
		log.Fatalf("Error fetching chain properties: %v", err)
	}
	fmt.Printf("✓ Account Creation Fee:    %s\n", props.AccountCreationFee)
	fmt.Printf("  Maximum Block Size:      %d bytes\n", props.MaximumBlockSize)
	fmt.Printf("  HBD Interest Rate:       %.2f%%\n", float64(props.HbdInterestRate)/100.0)
	fmt.Println()

	// 2. Account History Pagination
	fmt.Println("📖 [2] Fetching Account History Pagination...")
	accountName := "thecrazygm"
	// Fetch the last 5 operations
	history, err := api.GetAccountHistory(accountName, -1, 5)
	if err != nil {
		log.Fatalf("Error fetching account history: %v", err)
	}
	fmt.Printf("✓ Retrieved %d history items for @%s:\n", len(history), accountName)
	for _, item := range history {
		opName := ""
		if len(item.Op.Op) > 0 {
			opName, _ = item.Op.Op[0].(string)
		}
		fmt.Printf("  - [Seq %d] Block: %d | Op: %s\n", item.Seq, item.Op.Block, opName)
	}
	fmt.Println()

	// 3. Resource Credit & Voting Power Mana Math
	fmt.Println("⚡ [3] Resource Credit & Voting Power Mana Math...")
	accounts, err := api.GetAccounts([]string{accountName})
	if err != nil {
		log.Fatalf("Error fetching account details: %v", err)
	}
	if len(accounts) == 0 {
		log.Fatalf("Account @%s not found", accountName)
	}
	accData := accounts[0]

	// Calculate current real-time voting power (handling linear regeneration)
	vp := api.CalculateVPMana(accData)
	fmt.Printf("✓ Current Regenerated Voting Power:  %.2f%%\n", vp)

	// Fetch RC Details
	rcInfo, err := api.GetRCMana(accountName)
	if err != nil {
		log.Fatalf("Error fetching RC mana: %v", err)
	}
	fmt.Printf("✓ Max RC:                            %d\n", rcInfo.MaxMana)
	fmt.Printf("  Last Recorded RC:                  %d (%.2f%%)\n", rcInfo.LastMana, rcInfo.LastPercent)
	fmt.Printf("  Current Regenerated RC:            %d (%.2f%%)\n", rcInfo.CurrentMana, rcInfo.CurrentPercent)

	// Calculate current real-time RC using account data helper
	rcPercent := api.CalculateRCMana(accData)
	fmt.Printf("  Calculated RC via Helper:          %.2f%%\n", rcPercent)
	fmt.Println()

	fmt.Println("=== Examples Completed Successfully ===")
}
