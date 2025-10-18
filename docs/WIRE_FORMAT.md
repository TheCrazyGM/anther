# HIVE Wire Format: STEEM/SBD Legacy Support

## Critical Issue Resolved ✓

**Problem**: Signature validation was failing because HIVE transactions must use legacy STEEM/SBD symbols in the wire format for signing, even though users interact with HIVE/HBD.

**Solution**: Implemented automatic conversion from HIVE/HBD to STEEM/SBD during binary serialization for transaction signing.

## Background: The HIVE Fork

HIVE was forked from Steem in March 2020. The blockchain's transaction format still uses the legacy symbol names (STEEM, SBD) for wire protocol compatibility. However, modern interfaces present these as HIVE and HBD.

This is a long-standing compatibility issue with HIVE transactions.

## How It Works

### User Perspective (What You See)

```go
transfer := &transaction.Transfer{
    From:   "thecrazygm",
    To:     "ecoinstant",
    Amount: "0.001 HIVE",  // User-friendly HIVE
    Memo:   "Hello from golang!",
}
```

### Wire Format (What Gets Signed)

When the transaction is serialized for signing:

- `HIVE` → `STEEM` (in the wire format)
- `HBD` → `SBD` (in the wire format)

The conversion is automatic and transparent.

## Technical Implementation

### Wire Symbol Aliases

```go
var WireSymbolAliases = map[string]string{
	"HIVE": "STEEM",
	"HBD":  "SBD",
}
```

### Binary Serialization

The `Amount.Bytes()` method handles the conversion:

```
User Input: 0.001 HIVE
    ↓ Parse amount
    ↓ Get wire symbol (HIVE → STEEM)
    ↓ Serialize to binary

Wire Format:
┌─────────────────────┐
│ Amount (int64LE)    │ 0x0100000000000000 (1 satoshi)
├─────────────────────┤
│ Precision (uint8)   │ 0x03 (3 decimal places)
├─────────────────────┤
│ Symbol (7 bytes)    │ 0x535445454d0000 ("STEEM\0\0")
└─────────────────────┘
```

### Hex Breakdown

- Amount bytes: `01 00 00 00 00 00 00 00` = 1 (0.001 HIVE in satoshis)
- Precision: `03` = 3 decimal places
- Symbol: `53 54 45 45 4d 00 00` = "STEEM" (ASCII padded to 7 bytes)

## Asset Metadata

Supported assets and their properties:

```
HIVE   → STEEM   (precision: 3)   user sees HIVE
HBD    → SBD     (precision: 3)   user sees HBD
VESTS  → VESTS   (precision: 6)   no conversion needed
```

## Example: Transaction Signing Flow

### Step 1: Create Transfer

```go
transfer := &transaction.Transfer{
    From:   "sender",
    To:     "receiver",
    Amount: "0.001 HIVE",
    Memo:   "Payment",
}
```

### Step 2: Add to Transaction

```go
tx := transaction.NewTransaction(api)
tx.AppendOp(transfer)
```

### Step 3: Serialize for Signing

When `tx.Sign()` is called:

1. Transaction is converted to binary
2. `transfer.Bytes()` is called
3. Amount is parsed: 0.001 HIVE
4. Wire symbol conversion: HIVE → STEEM
5. Binary is created with "STEEM" in wire format
6. Signature is generated over the bytes

### Step 4: Broadcast

```go
result, err := tx.Broadcast()
```

The complete transaction with signature is sent to the network.

## Why This Matters

Without this conversion:

- Signatures wouldn't match blockchain expectations
- Transaction signing would fail silently
- Transfers would be rejected by the network

With this conversion:

- Signatures are valid according to HIVE blockchain rules
- Transactions are accepted by the network
- Everything works seamlessly for the user

## User Experience

Users should **never** need to know about STEEM/SBD:

```go
// ✓ CORRECT - User-friendly
amount := "0.001 HIVE"
amount := "1.500 HBD"

// ✗ NOT NEEDED - Happens automatically
amount := "0.001 STEEM"  // Wrong! Use HIVE instead
amount := "1.500 SBD"    // Wrong! Use HBD instead
```

The library handles the conversion transparently during signing.

## Implementation Details

### Types Package

The `Amount` type in `types/types.go`:

- Parses user-friendly amounts ("0.001 HIVE")
- Converts to binary with wire symbols
- Handles precision (decimal places)
- Pads symbols to 7 bytes in wire format

### Transaction Package

The `Transfer` type in `transaction/transaction.go`:

- Uses `Amount.Bytes()` for serialization
- Calls wire format conversion automatically
- No special handling needed by user

## Code Example: Complete Flow

```go
package main

import (
    "github.com/thecrazygm/nectar-go/client"
    "github.com/thecrazygm/nectar-go/transaction"
    "github.com/thecrazygm/nectar-go/types"
    "github.com/thecrazygm/nectar-go/wallet"
)

func main() {
    api := client.NewClient([]string{"https://api.hive.blog"}, 30)

    // Create transfer - user sees HIVE
    tx := transaction.NewTransaction(api)
    tx.AppendOp(&transaction.Transfer{
        From:   "sender",
        To:     "receiver",
        Amount: "0.001 HIVE",  // User input
        Memo:   "Payment",
    })

    // Sign - automatic HIVE→STEEM conversion happens here
    w := wallet.NewWallet()
    w.AddKey("sender", "active", "5KxxxWIF")
    w.Sign(tx, "sender", "active")  // Wire format uses STEEM internally

    // Broadcast
    result, _ := tx.Broadcast()
}
```

## Testing

To verify the wire format conversion:

```bash
# Build the test
go build examples/

# Run a transfer with valid WIF
export ACTIVE_WIF="5Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
./examples/transfer
```

If the signatures match blockchain expectations, the transaction will be accepted.

## Compatibility

This implementation matches:

- ✓ Python nectarlite library behavior
- ✓ Hive blockchain wire protocol
- ✓ Legacy STEEM transaction format
- ✓ Modern HIVE/HBD user interface

## References

- [HIVE Fork History](https://hive.io/)
- [Blockchain Transaction Format](https://hiveblocks.com/)
- [Wire Protocol Documentation](https://developers.hive.io/)

## Summary

The HIVE blockchain requires STEEM/SBD in the wire format for historical compatibility. This library:

1. **Accepts** user input as HIVE/HBD (modern names)
2. **Converts** to STEEM/SBD during binary serialization
3. **Signs** with the legacy wire format
4. **Broadcasts** valid transactions to the network

Users never see or need to input STEEM/SBD - it's handled automatically! ✓
