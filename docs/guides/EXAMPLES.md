# Anther Go - Usage Examples

This document demonstrates how to use the Anther Go library for common Hive blockchain operations.

## Running the Examples

### 1. Account Query Example (main.go)

Query account information, voting power, and resource credits:

```bash
./nectar-go
```

**Output includes:**

- Account basic information (name, balance, vesting shares)
- Current voting power with mana regeneration
- Resource Credit (RC) information
- Transfer transaction example (unsigned)

**Example Output:**

```text
=== Anther Go Library - Account Query ===

Fetching account data for 'thecrazygm'...
✓ Account data fetched successfully

--- Basic Account Information ---
Account Name: thecrazygm
Balance: 69.997 HIVE
HBD Balance: 0.000 HBD
Vesting Shares: 10042034.757156 VESTS
Created: 2017-05-05T23:09:12

--- Voting Power ---
Current Voting Power: 55.03%

--- Resource Credits (RC) ---
Current RC: 100.00%
Last RC: 100.00%
Current Mana: 10044055506129
```

### 2. Transfer Example (examples/transfer.go)

Send a 0.001 HIVE transfer with a memo, signed with an active key:

```bash
# Build the transfer example
go build -o examples/transfer ./examples

# Run with your active WIF key from environment
export ACTIVE_WIF="5Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
./examples/transfer
```

**What it does:**

1. Loads the active key from the `ACTIVE_WIF` environment variable
2. Creates a wallet and adds the key
3. Fetches account data
4. Creates a transfer transaction:
   - From: thecrazygm
   - To: ecoinstant
   - Amount: 0.001 HIVE
   - Memo: "Hello from golang!"
5. Signs the transaction with the active key
6. Broadcasts it to the network

**Example Output:**

```text
=== Anther Go Library - Transfer Example ===

Adding active key to wallet...
✓ Key added successfully

Fetching account data...
✓ Account data fetched successfully

--- Account Information ---
Account: thecrazygm
Current Balance: 69.997 HIVE

--- Creating Transfer ---
Transfer prepared:
  From: thecrazygm
  To: ecoinstant
  Amount: 0.001 HIVE
  Memo: Hello from golang!

--- Signing Transaction ---
Signing with active key...
✓ Transaction signed successfully
✓ Signatures: 1

--- Broadcasting Transaction ---
Broadcasting to network...
✓ Transaction broadcast successfully!

Result: <transaction_response>

=== Transfer Complete ===
```

## Library Usage

### Basic Account Operations

```go
package main

import (
    "github.com/srbde/hive-anther/account"
    "github.com/srbde/hive-anther/client"
)

func main() {
    // Create API client
    api := client.NewClient([]string{"https://api.hive.blog"}, 30)

    // Create account instance
    acc := account.NewAccount("username", api)

    // Fetch account data
    acc.Refresh()

    // Get voting power
    vp, _ := acc.VotingPower()

    // Get resource credits
    rc, _ := acc.RC()

    // Get RC info
    rcInfo, _ := acc.RCInfo()
}
```

### Transfer Operation

```go
package main

import (
    "github.com/srbde/hive-anther/transaction"
    "github.com/srbde/hive-anther/wallet"
    "github.com/srbde/hive-anther/client"
)

func main() {
    api := client.NewClient([]string{"https://api.hive.blog"}, 30)

    // Create transaction
    tx := transaction.NewTransaction(api)

    // Create transfer operation
    transfer := &transaction.Transfer{
        From:   "sender_account",
        To:     "receiver_account",
        Amount: "1.000 HIVE",
        Memo:   "Payment for services",
    }

    tx.AppendOp(transfer)

    // Sign with wallet
    w := wallet.NewWallet()
    w.AddKey("sender_account", "active", "5KxxxWIFKey")
    w.Sign(tx, "sender_account", "active")

    // Broadcast
    result, _ := tx.Broadcast()
}
```

### Vote Operation

```go
package main

import (
    "github.com/srbde/hive-anther/transaction"
    "github.com/srbde/hive-anther/wallet"
    "github.com/srbde/hive-anther/client"
)

func main() {
    api := client.NewClient([]string{"https://api.hive.blog"}, 30)

    tx := transaction.NewTransaction(api)

    // Create vote operation (weight is in basis points, 10000 = 100%)
    vote := &transaction.Vote{
        Voter:    "your_account",
        Author:   "post_author",
        Permlink: "post-slug",
        Weight:   10000, // 100% upvote
    }

    tx.AppendOp(vote)

    w := wallet.NewWallet()
    w.AddKey("your_account", "posting", "5KxxxWIFKey")
    w.Sign(tx, "your_account", "posting")

    result, _ := tx.Broadcast()
}
```

### Follow Operation

```go
package main

import (
    "github.com/srbde/hive-anther/account"
    "github.com/srbde/hive-anther/wallet"
    "github.com/srbde/hive-anther/client"
)

func main() {
    api := client.NewClient([]string{"https://api.hive.blog"}, 30)

    acc := account.NewAccount("your_account", api)

    // Create follow transaction
    tx, _ := acc.Follow("account_to_follow")

    // Sign and broadcast
    w := wallet.NewWallet()
    w.AddKey("your_account", "posting", "5KxxxWIFKey")
    w.Sign(tx, "your_account", "posting")

    result, _ := tx.Broadcast()
}
```

## Environment Variables

The transfer example uses the `ACTIVE_WIF` environment variable:

```bash
export ACTIVE_WIF="5Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

**Never commit your WIF keys to version control!** Use environment variables or secure configuration files.

## Error Handling

The library provides custom exception types:

```go
import "github.com/srbde/hive-anther/exceptions"

// Transaction errors
err := tx.Sign(wif)
if err != nil {
    // Handle transaction error
}

// Wallet errors
err := wallet.AddKey("account", "active", wif)
if err != nil {
    // Handle key validation error (InvalidKeyFormatError, etc.)
}

// API errors
_, err := api.Call("condenser_api", "get_accounts", params)
if err != nil {
    // Handle node error
}
```

## Key Features

- ✅ Multi-node failover support
- ✅ Account data retrieval and caching
- ✅ Voting power calculation with mana regeneration
- ✅ Resource Credit tracking
- ✅ Private key management with WIF validation
- ✅ Transaction signing with ECDSA
- ✅ Multiple operation types (Vote, Transfer, Comment, Follow, CustomJSON)
- ✅ Proper error handling with custom exceptions

## Building Examples

Build all examples:

```bash
go build -o examples/transfer ./examples
```

Build main query tool:

```bash
go build -o nectar-go
```

## API Endpoints

Anther connects to public Hive nodes:

- <https://api.hive.blog> (primary)
- <https://api.syncad.com> (fallback)

You can specify custom nodes when creating a client:

```go
api := client.NewClient([]string{
    "https://custom-node-1.hive.com",
    "https://custom-node-2.hive.com",
}, timeout)
```

## More Information

For more details on the Hive blockchain, visit:

- <https://developers.hive.io/>
- <https://hiveblocks.com/>
- <https://peakd.com/>

For the Python reference implementation, see:

- <https://github.com/hivecommunity/anther>
