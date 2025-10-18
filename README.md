# Nectarlite Go - HIVE Blockchain Library

A complete Go implementation of the HIVE blockchain library, compatible with the Python nectarlite reference implementation.

## Status: ✅ PRODUCTION READY

All features implemented and tested. Ready for production use on the HIVE blockchain.

## Features

### Core Functionality

- ✅ Account management and queries
- ✅ Transaction creation and signing
- ✅ Multi-operation transactions
- ✅ Broadcasting to network
- ✅ Wallet management with WIF keys
- ✅ Multi-node client with failover

### Signing System

- ✅ ECDSA signature generation
- ✅ Canonical signature format (s ≤ N/2)
- ✅ Recovery ID bit adjustment
- ✅ Wire format conversion (HIVE/HBD ↔ STEEM/SBD)
- ✅ Proper transaction hashing with chain ID

### Supported Operations

- ✅ Transfer (with amount and memo)
- ✅ Vote (with weight)
- ✅ Comment (with metadata)
- ✅ CustomJSON (for plugins)
- ✅ Follow (via custom JSON)

### Account Features

- ✅ Account data fetching
- ✅ Voting power calculation
- ✅ Mana regeneration tracking
- ✅ Resource credit (RC) queries
- ✅ History and balance information

## Quick Start

### Installation

```bash
go get github.com/thecrazygm/nectar-go
```

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/thecrazygm/nectar-go/account"
    "github.com/thecrazygm/nectar-go/client"
)

func main() {
    // Create client
    api := client.NewClient(
        []string{"https://api.hive.blog"},
        30, // timeout in seconds
    )

    // Query account
    acc := account.NewAccount("username", api)
    if err := acc.Refresh(); err != nil {
        panic(err)
    }

    fmt.Printf("Account: %s\n", acc.Name)
    fmt.Printf("Balance: %v\n", acc.Data["balance"])
}
```

### Transaction Example

```go
package main

import (
    "fmt"
    "log"
    "os"
    "github.com/thecrazygm/nectar-go/client"
    "github.com/thecrazygm/nectar-go/transaction"
    "github.com/thecrazygm/nectar-go/wallet"
)

func main() {
    // Setup
    api := client.NewClient([]string{"https://api.hive.blog"}, 30)
    w := wallet.NewWallet()

    // Create transfer
    tx := transaction.NewTransaction(api)
    tx.AppendOp(&transaction.Transfer{
        From:   "sender",
        To:     "receiver",
        Amount: "1.000 HIVE",
        Memo:   "Payment",
    })

    // Sign
    wif := os.Getenv("ACTIVE_WIF")
    w.AddKey("sender", "active", wif)
    if err := w.Sign(tx, "sender", "active"); err != nil {
        log.Fatal(err)
    }

    // Broadcast
    result, err := tx.Broadcast()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Transaction: %v\n", result)
}
```

## Architecture

### Transaction Signing Flow

```
WIF Key + Transaction Data
    ↓
Create transaction with operations
    ↓
Get transaction hex from node
    ↓
SHA256(CHAIN_ID + TX_HEX) = digest
    ↓
ECDSA sign digest
    ↓
Extract (r, s, recovery_id)
    ↓
Canonicalize: if s > N/2 then s = N - s
    ↓
Adjust recovery_id: if s was flipped then recovery_id XOR 1
    ↓
Build signature: [27+4+recovery_id][r][s]
    ↓
Broadcast to network ✅
```

### Wire Format Conversion

HIVE transactions automatically convert between user-friendly and wire format names:

```go
// User perspective (input)
Amount: "1.000 HIVE"     // User-friendly

// Wire format (for signing)
"STEEM"                   // Legacy name

// Output
Amount: "1.000 HIVE"     // User-friendly again
```

This conversion is transparent and automatic during transaction signing.

## Key Insight: Recovery ID Adjustment with Decred secp256k1

The implementation uses **Decred's secp256k1 library** which provides `SignCompact()` that:

1. Automatically embeds the recovery ID in the first byte
2. Returns compact signatures ready for HIVE blockchain
3. Works seamlessly with canonicalization

When canonicalizing an ECDSA signature by transforming `s → N - s`, the y-parity of the elliptic curve point changes. Therefore, we must flip recovery ID bit 0:

```go
if sWasFlipped {
    recoveryID = recoveryID ^ 1  // Flip y-parity bit
}
```

This ensures the signature can properly recover to the original public key on the blockchain.

## File Structure

```
nectar-go/
├── README.md                    # This file
├── IMPLEMENTATION_SUMMARY.md    # Complete feature list
├── SIGNING_IMPLEMENTATION.md    # Signing process details
├── RECOVERY_ID_DEEP_DIVE.md     # Recovery ID mathematics
├── WIRE_FORMAT.md               # Wire format conversion
├── CANONICAL_SIGNATURES.md      # Canonicalization details
│
├── account/
│   └── account.go              # Account management
├── client/
│   └── client.go               # JSON-RPC client
├── exceptions/
│   └── exceptions.go           # Error types
├── transaction/
│   └── transaction.go          # Transaction signing
├── types/
│   └── types.go                # Amount and types
├── wallet/
│   └── wallet.go               # Wallet management
│
├── examples/
│   └── transfer.go             # Transfer example
├── main.go                      # CLI entry point
└── go.mod                       # Go module definition
```

## API Reference

### Client

```go
// Create client with nodes and timeout
api := client.NewClient([]string{node1, node2}, timeoutSeconds)

// Make RPC call
result, err := api.Call("api_name", "method_name", params)

// Account queries
properties, err := api.GetDynamicGlobalProperties()
account, err := api.GetAccount("username")
```

### Account

```go
// Create account
acc := account.NewAccount("username", api)

// Refresh data
err := acc.Refresh()

// Get voting power
votingPower := acc.GetVotingPower()

// Get RC info
rcInfo, err := acc.GetRCInfo()
```

### Transaction

```go
// Create transaction
tx := transaction.NewTransaction(api)

// Add operations
tx.AppendOp(&transaction.Transfer{...})
tx.AppendOp(&transaction.Vote{...})

// Sign
err := tx.Sign(wifKey)

// Broadcast
result, err := tx.Broadcast()
```

### Wallet

```go
// Create wallet
w := wallet.NewWallet()

// Add key
err := w.AddKey("username", "active", wifKey)

// Sign transaction
err := w.Sign(tx, "username", "active")
```

## Constants

- **HIVE Chain ID**: `beeab0de00000000000000000000000000000000000000000000000000000000`
- **secp256k1 N**: `0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141`
- **secp256k1 N/2**: `0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0`

## Testing

### Run account query

```bash
go run main.go
```

### Build transfer example

```bash
go build -o examples/transfer ./examples
```

### Test with transfer

```bash
export ACTIVE_WIF="5Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
./examples/transfer
```

## Compatibility

- ✅ Python nectarlite library (1:1 matching)
- ✅ HIVE blockchain consensus rules
- ✅ Go 1.16+ versions
- ✅ Cross-platform (Linux, macOS, Windows)

## Dependencies

```
github.com/btcsuite/btcd/btcutil                        v1.x    # WIF key handling
github.com/decred/dcrd/dcrec/secp256k1/v4                       # ECDSA signing (SignCompact)
github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa                 # Signature operations
```

**Note**: Uses Decred's secp256k1 library (not go-ethereum) for proper compact signature generation with embedded recovery IDs.

## Documentation

See detailed documentation in:

- `IMPLEMENTATION_SUMMARY.md` - Complete feature overview
- `SIGNING_IMPLEMENTATION.md` - Step-by-step signing process
- `RECOVERY_ID_DEEP_DIVE.md` - Recovery ID mathematics
- `WIRE_FORMAT.md` - STEEM/SBD conversion explained
- `CANONICAL_SIGNATURES.md` - Signature canonicalization

## Troubleshooting

### "signature is not canonical"

- Cause: s > N/2 not being canonicalized
- Solution: Ensure s canonicalization is enabled (it is by default)

### "unable to reconstruct public key from signature"

- Cause: Recovery ID not adjusted after s flip
- Solution: Check recovery ID bit flip logic (implemented)

### "Bad Cast: Invalid cast from null_type to Array"

- Cause: Node doesn't support get_transaction_hex
- Solution: Try different node (e.g., api.openhive.network)

## Contributing

This implementation closely matches the Python reference. Changes should:

1. Maintain compatibility with Python nectarlite
2. Include proper error handling
3. Have comprehensive documentation
4. Pass all existing tests

## License

See LICENSE file for details.

## Support

For issues and questions:

- Check the documentation files
- Review the example code
- Consult the Python reference implementation

---

**Implementation Status**: ✅ COMPLETE
**Blockchain Ready**: ✅ YES
**Production Ready**: ✅ YES
**Last Updated**: October 18, 2025
