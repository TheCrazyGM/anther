# HIVE Transaction Signing Implementation

## Overview

This document describes the complete transaction signing process for the HIVE blockchain in Go, including wire format conversion, signature canonicalization, and recovery ID handling.

## Signing Flow

### Phase 1: Transaction Preparation
1. Create transaction with operations (transfer, vote, comment, etc.)
2. Operations are converted to wire format with STEEM/SBD conversion
3. Get transaction hex from the node via RPC

### Phase 2: Message Hashing
1. Chain ID + transaction hex → concatenate
2. SHA256 hash of the concatenated data
3. This is the digest that will be signed

### Phase 3: ECDSA Signing
```
digest (SHA256) → ECDSA sign with private key → 65-byte signature
  [recovery_byte][r: 32 bytes][s: 32 bytes]
```

### Phase 4: S Canonicalization

HIVE requires canonical signatures where `s ≤ N/2` (where N is the secp256k1 curve order).

```
If s > N/2:
    s_canonical = N - s
Else:
    s_canonical = s
```

This ensures:
- Non-malleability of signatures
- Deterministic signature format
- Blockchain consensus compliance

### Phase 5: Recovery ID Adjustment

When `s` is canonicalized, the recovery formula changes. We must adjust the recovery ID accordingly:

**Recovery ID Encoding (0-3):**
- **Bit 0 (recovery_id % 2)**: Y-coordinate parity (0=even, 1=odd)
- **Bit 1 (recovery_id / 2)**: X-coordinate overflow (0=no overflow, 1=overflow)

**When s is flipped during canonicalization:**
- The y-parity changes
- Therefore, we **flip bit 0** of the recovery ID
- `new_recovery_id = old_recovery_id ^ 1`

**When s is already canonical:**
- Recovery ID remains unchanged

### Phase 6: Final Signature

```
Final Signature = [27 + 4 + recovery_id][r: 32 bytes][s: 32 bytes]
                = [31-34][32 bytes][32 bytes]
                = 65 bytes total
```

Where:
- **27**: Bitcoin's base recovery byte value
- **4**: Compressed public key indicator
- **recovery_id**: 0-3, adjusted if s was canonicalized

## Wire Format: STEEM/SBD Conversion

### User Perspective
```go
transfer := &transaction.Transfer{
    From:   "alice",
    To:     "bob",
    Amount: "1.000 HIVE",  // User-friendly
    Memo:   "Payment",
}
```

### Wire Format (for signing)
- HIVE → STEEM (during binary serialization)
- HBD → SBD (during binary serialization)
- This conversion is automatic and transparent

**Why:** HIVE forked from Steem in 2020 but kept legacy wire protocol names for blockchain compatibility.

## Code Implementation

### File: transaction/transaction.go - Sign() method

```go
// Step 1: Sign with private key
sig, err := crypto.Sign(digest[:], privKey)

// Step 2: Extract components
recoveryByteFromGo := sig[0]
recoveryIDFromGo := int((recoveryByteFromGo - 27) % 4)  // Extract 0-3
rBytes := sig[1:33]
sBytes := sig[33:65]

// Step 3: Parse s value
s := new(big.Int).SetBytes(sBytes)

// Step 4: Check and canonicalize s
nDiv2 := new(big.Int)
nDiv2.SetString("7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0", 16)

sNeedsFlip := false
if s.Cmp(nDiv2) > 0 {
    // Canonicalize: s = N - s
    nValue := new(big.Int)
    nValue.SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
    s = new(big.Int).Sub(nValue, s)
    sNeedsFlip = true
}

// Step 5: Reconstruct as 32-byte values
sBytes = s.Bytes()
if len(sBytes) < 32 {
    sBytes = append(make([]byte, 32-len(sBytes)), sBytes...)
}

// Step 6: Build canonical signature
canonical := append(rBytes, sBytes...)

// Step 7: Adjust recovery ID if needed
recoveryID := recoveryIDFromGo
if sNeedsFlip {
    recoveryID = recoveryID ^ 1  // Flip y-parity bit
}

// Step 8: Create final signature
finalSig := append([]byte{byte(27 + 4 + recoveryID)}, canonical...)
tx.Signatures = append(tx.Signatures, hex.EncodeToString(finalSig))
```

## Constants

### secp256k1 Curve Parameters
```
N (curve order) = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
N/2             = 0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0
```

### HIVE Chain ID
```
HIVE_CHAIN_ID = "beeab0de00000000000000000000000000000000000000000000000000000000"
```

## Wire Format Conversion

### File: types/types.go

```go
var WireSymbolAliases = map[string]string{
    "HIVE": "STEEM",
    "HBD":  "SBD",
}

// Amount.Bytes() method:
// 1. Parse user amount (e.g., "1.000 HIVE")
// 2. Get wire symbol (HIVE → STEEM)
// 3. Convert amount to satoshis (int64)
// 4. Binary format: [amount_int64LE][precision_uint8][symbol_7bytes_padded]
```

## Testing

### Build the transfer example:
```bash
go build -o examples/transfer ./examples
```

### Test with valid WIF:
```bash
export ACTIVE_WIF="5Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
./examples/transfer
```

### What happens:
1. Account queries work ✓
2. Transaction signing produces canonical signature ✓
3. Recovery ID is correctly adjusted if s was canonicalized ✓
4. Wire format uses STEEM/SBD internally ✓
5. Transaction broadcasts to network ✓

## Validation Checklist

- [x] Signature is canonical (s ≤ N/2)
- [x] Recovery ID is adjusted when s is flipped
- [x] Recovery ID is 0-3 (compressed key format)
- [x] Wire format converts HIVE/HBD to STEEM/SBD
- [x] Transaction hex includes all operations
- [x] Chain ID is included in digest
- [x] Signature is 65 bytes: [1 recovery byte][32 r][32 s]

## Common Issues and Fixes

### "signature is not canonical"
- **Cause**: s > N/2 and wasn't canonicalized
- **Fix**: Check `if s > N/2: s = N - s` is executed

### "unable to reconstruct public key from signature"
- **Cause**: Recovery ID is wrong, likely wasn't adjusted after s flip
- **Fix**: Flip recovery ID bit 0 when s is canonicalized

### "Bad Cast: Invalid cast from null_type to Array"
- **Cause**: API node doesn't support get_transaction_hex
- **Fix**: Try different node, e.g., api.openhive.network

## References

- [Bitcoin BIP-62: DER Signature Canonicalization](https://github.com/bitcoin/bips/blob/master/bip-0062.md)
- [secp256k1 Curve Parameters](https://en.wikipedia.org/wiki/Elliptic_Curve_Digital_Signature_Algorithm)
- [HIVE Blockchain Consensus Rules](https://developers.hive.io/)
- [go-ethereum Crypto Package](https://github.com/ethereum/go-ethereum/tree/master/crypto)

## Summary

The HIVE transaction signing process ensures:
1. **Deterministic signatures** through s canonicalization
2. **Proper recovery** through recovery ID bit adjustment
3. **Blockchain compatibility** through wire format conversion
4. **Non-malleability** through canonical form requirement

All of these are handled automatically when calling `tx.Sign(wif)`.
