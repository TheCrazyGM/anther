package main

import (
	"fmt"
	"log"

	"github.com/thecrazygm/nectar-go/account"
	"github.com/thecrazygm/nectar-go/client"
	"github.com/thecrazygm/nectar-go/transaction"
)

func main() {
	// Initialize the client with Hive nodes
	nodes := []string{
		"https://api.hive.blog",
		"https://api.hivecosystem.dev",
	}
	api := client.NewClient(nodes, 30)

	// Create an account instance for "thecrazygm"
	acc := account.NewAccount("thecrazygm", api)

	fmt.Println("=== Nectarlite Go Library - Account Query ===\n")

	// Refresh account data
	fmt.Println("Fetching account data for 'thecrazygm'...")
	if err := acc.Refresh(); err != nil {
		log.Fatalf("Error refreshing account: %v", err)
	}
	fmt.Println("✓ Account data fetched successfully\n")

	// Display basic account info
	fmt.Println("--- Basic Account Information ---")
	fmt.Printf("Account Name: %s\n", acc.Name)
	if balance, ok := acc.Data["balance"].(string); ok {
		fmt.Printf("Balance: %s\n", balance)
	}
	if hbdBalance, ok := acc.Data["hbd_balance"].(string); ok {
		fmt.Printf("HBD Balance: %s\n", hbdBalance)
	}
	if vestingShares, ok := acc.Data["vesting_shares"].(string); ok {
		fmt.Printf("Vesting Shares: %s\n", vestingShares)
	}
	if createdAt, ok := acc.Data["created"].(string); ok {
		fmt.Printf("Created: %s\n", createdAt)
	}

	// Get voting power
	fmt.Println("\n--- Voting Power ---")
	vp, err := acc.VotingPower()
	if err != nil {
		log.Printf("Error getting voting power: %v", err)
	} else {
		fmt.Printf("Current Voting Power: %.2f%%\n", vp)
	}

	// Get RC info
	fmt.Println("\n--- Resource Credits (RC) ---")
	rcInfo, err := acc.RCInfo()
	if err != nil {
		log.Printf("Note: Could not fetch RC info: %v", err)
	} else {
		if currentPercent, ok := rcInfo["current_percent"].(float64); ok {
			fmt.Printf("Current RC: %.2f%%\n", currentPercent)
		}
		if lastPercent, ok := rcInfo["last_percent"].(float64); ok {
			fmt.Printf("Last RC: %.2f%%\n", lastPercent)
		}
		if currentMana, ok := rcInfo["current_mana"].(int64); ok {
			fmt.Printf("Current Mana: %d\n", currentMana)
		}
	}

	// Demonstrate creating a transfer transaction (without actual broadcast)
	fmt.Println("\n--- Transfer Example (Unsigned) ---")
	tx := transaction.NewTransaction(api)
	transfer := &transaction.Transfer{
		From:   "thecrazygm",
		To:     "ecoinstant",
		Amount: "0.001 HIVE",
		Memo:   "Hello from golang!",
	}
	tx.AppendOp(transfer)
	fmt.Printf("Transfer prepared (ready to sign with ACTIVE_WIF):\n")
	fmt.Printf("  From: %s\n", transfer.From)
	fmt.Printf("  To: %s\n", transfer.To)
	fmt.Printf("  Amount: %s\n", transfer.Amount)
	fmt.Printf("  Memo: %s\n", transfer.Memo)
	fmt.Printf("\nRun the transfer example with ACTIVE_WIF env var to broadcast:\n")
	fmt.Printf("  ACTIVE_WIF=<your-key> ./examples/transfer\n")

	fmt.Println("\n=== Query Complete ===")
}
