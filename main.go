package main

import (
	"fmt"
	"log"
	"os"

	"github.com/thecrazygm/nectar-go/account"
	"github.com/thecrazygm/nectar-go/client"
	"github.com/thecrazygm/nectar-go/transaction"
	"github.com/thecrazygm/nectar-go/wallet"
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

	fmt.Println("=== Nectarlite Go Library - Account Query ===")

	// Refresh account data
	fmt.Println("Fetching account data for 'thecrazygm'...")
	if err := acc.Refresh(); err != nil {
		log.Fatalf("Error refreshing account: %v", err)
	}
	fmt.Println("✓ Account data fetched successfully")

	// Display basic account info
	fmt.Println("--- Basic Account Information ---")
	fmt.Printf("Account Name: %s\n", acc.Name)
	if reputation, err := acc.Reputation(); err != nil {
		log.Printf("Error getting reputation: %v", err)
	} else {
		fmt.Printf("Reputation: %d\n", reputation)
	}
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
	activeWIF := os.Getenv("ACTIVE_WIF")
	if activeWIF == "" {
		fmt.Println("\n--- Signing Skipped ---")
		fmt.Println("ACTIVE_WIF not set; unable to sign sample transfer.")
	} else {
		fmt.Println("\n--- Signing Transfer ---")
		w := wallet.NewWallet()
		if err := w.AddKey("thecrazygm", "active", activeWIF); err != nil {
			log.Printf("Error adding key to wallet: %v", err)
		} else if err := w.Sign(tx, "thecrazygm", "active"); err != nil {
			log.Printf("Error signing transfer: %v", err)
		} else if len(tx.Signatures) > 0 {
			fmt.Printf("Signature: %s\n", tx.Signatures[0])
		}
	}

	fmt.Printf("\nRun the transfer example with ACTIVE_WIF env var to broadcast:\n")
	fmt.Printf("  ACTIVE_WIF=<your-key> ./examples/transfer\n")

	fmt.Println("\n=== Query Complete ===")
}
