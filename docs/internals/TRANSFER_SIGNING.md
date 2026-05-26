# Transfer Signing Guide

## Overview

The Anther Go library supports creating and signing transfer transactions on the Hive blockchain. This guide explains the signing process and how to use it.

## Transfer Example

### Basic Transfer

```go
package main

import (
    "github.com/srbde/anther/account"
    "github.com/srbde/anther/client"
    "github.com/srbde/anther/transaction"
    "github.com/srbde/anther/wallet"
)

func main() {
    // Initialize API client
    api := client.NewClient([]string{"https://api.hive.blog"}, 30)

    // Create transaction
    tx := transaction.NewTransaction(api)

    // Create transfer operation
    transfer := &transaction.Transfer{
        From:   "sender_account",
        To:     "receiver_account",
        Amount: "0.001 HIVE",
        Memo:   "Hello from golang!",
    }

    // Add operation to transaction
    tx.AppendOp(transfer)

    // Create and configure wallet
    w := wallet.NewWallet()
    w.AddKey("sender_account", "active", "5KxxxWIFKey")

    // Sign transaction
    if err := w.Sign(tx, "sender_account", "active"); err != nil {
        panic(err)
    }

    // Broadcast
    result, err := tx.Broadcast()
    if err != nil {
        panic(err)
    }
}
```

## Transaction Signing Process

### Step 1: Transaction Creation

```go
tx := transaction.NewTransaction(api)
```

- Creates an empty transaction with reference to the API
- Initializes empty operations and signatures lists

### Step 2: Add Operations

```go
transfer := &transaction.Transfer{
    From:   "thecrazygm",
    To:     "ecoinstant",
    Amount: "0.001 HIVE",
    Memo:   "Hello from golang!",
}
tx.AppendOp(transfer)
```

- Transfer objects contain the operation details
- Operations implement the `Operation` interface with `ToDict()` method

### Step 3: Set Block Parameters

When signing, the transaction automatically:

- Fetches the latest block number from the network
- Calculates reference block number and prefix
- Sets expiration time (30 seconds from block time)

```go
// This happens automatically in tx.Sign()
props, err := api.GetDynamicGlobalProperties()
// Sets: tx.RefBlockNum, tx.RefBlockPrefix, tx.Expiration
```

### Step 4: Construct Transaction Dictionary

The transaction is converted to a dictionary format:

```go
{
  "ref_block_num": 1234,
  "ref_block_prefix": 5678,
  "expiration": "2025-10-18T08:00:00",
  "operations": [
    [
      "transfer",
      {
        "from": "thecrazygm",
        "to": "ecoinstant",
        "amount": "0.001 HIVE",
        "memo": "Hello from golang!"
      }
    ]
  ],
  "extensions": [],
  "signatures": []
}
```

### Step 5: Get Transaction Hex

```go
// API call to get the serialized hex
txHex = api.Call("condenser_api", "get_transaction_hex", [tx_dict])
// Returns: "d2042e1600000049f36801020a..."
```

### Step 6: Sign Digest

```go
// Prepare message: chain_id + transaction_hex
message = bytes.fromhex(HIVE_CHAIN_ID + txHex[:-2])
digest = sha256(message)

// Sign with private key (ECDSA)
signature = sign(digest, private_key)
```

### Step 7: Add Signature to Transaction

```go
tx.Signatures = append(tx.Signatures, signature_hex)
```

### Step 8: Broadcast

```go
result, err := tx.Broadcast()
```

- Sends the signed transaction to the network
- Returns transaction ID or error

## Common Issues and Solutions

### Issue 1: "Invalid WIF Format"

**Problem**: The private key provided is not in valid WIF format

**Solution**:

- Ensure the key starts with "5"
- Verify the key is not corrupted
- Check it's not accidentally truncated

```bash
export ACTIVE_WIF="5Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

### Issue 2: "Error calling get_transaction_hex"

**Problem**: The `get_transaction_hex` API call failed

**Possible Causes**:

- Node doesn't support this API method
- Node is in maintenance
- Network connection issue
- Invalid transaction data

**Solution**:

```go
// Retry with different nodes
nodes := []string{
    "https://api.hive.blog",
    "https://api.hivecosystem.dev",
    "https://anyx.io",
}
api := client.NewClient(nodes, 30)
```

### Issue 3: "Transaction is not signed"

**Problem**: Attempting to broadcast without signatures

**Solution**:

```go
// Ensure you call wallet.Sign() before Broadcast()
if err := w.Sign(tx, account, role); err != nil {
    return err
}
// Only then:
result, err := tx.Broadcast()
```

### Issue 4: "Unexpected response type"

**Problem**: API returned unexpected data type

**Possible Causes**:

- API endpoint changed
- Network returned error wrapped in JSON
- Response parsing issue

**Solution**:

- Check error message for details
- Verify you're using correct API endpoint
- Try alternate nodes

## Transfer Amount Format

Amounts must be strings with denomination:

- `"0.001 HIVE"` - 1 HIVE satoshi
- `"1.000 HIVE"` - 1 whole HIVE
- `"100.000 HIVE"` - 100 HIVE
- `"0.001 HBD"` - HBD transfers

Memo:

- Maximum 256 characters
- Optional (can be empty string)
- Appears on blockchain for both sender and receiver

## Security Notes

⚠️ **NEVER hardcode private keys!**

✅ Best Practices:

```bash
# Use environment variables
export ACTIVE_WIF="5Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# Load from secure config
source ~/.hive/keys.sh

# Use key management service (production)
```

## Running the Transfer Example

### Prerequisites

1. Valid Hive account name
2. Active key in WIF format
3. Sufficient HIVE balance for transfer + resource credits

### Execution

```bash
# Build the example
go build -o examples/transfer ./examples

# Set your active key
export ACTIVE_WIF="5Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# Run the transfer
./examples/transfer
```

### Expected Output

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

## Verifying Your Transfer

After successful broadcast, verify the transaction:

1. **Hive Blocks (Explorer)**
   - <https://hiveblocks.com/>
   - Search for your transaction ID

2. **PeakD (Community UI)**
   - <https://peakd.com/>
   - View in account transactions

3. **Command Line**

   ```bash
   # Check account balance
   ./nectar-go
   ```

## Transaction Fees

Each transfer costs:

- **RC (Resource Credits)**: ~50-100 RC depending on memo length
- **No direct HIVE fee**: Paid via resource credits that regenerate
- **Mana regenerates**: Over 5 days (432000 seconds)

## API Methods Used

- `condenser_api.get_dynamic_global_properties` - Get block parameters
- `block_api.get_block` - Get block details
- `condenser_api.get_transaction_hex` - Serialize transaction
- `condenser_api.broadcast_transaction_synchronous` - Broadcast transaction

## References

- [Hive Developer Docs](https://developers.hive.io/)
- [Transaction Format](https://hiveblocks.com/)
- [JSON-RPC API](https://developers.hive.io/apis/)
- [Hive Blockchain Explorer](https://hiveblocks.com/)
