# Anther Go - Quick Reference Guide

## Critical Code: Signature Signing with Recovery ID Adjustment

### Location: `transaction/transaction.go` - Sign() method

```go
// Step 1: Convert WIF to secp256k1 private key
wifDecoded, _ := btcutil.DecodeWIF(wif)
privKeyBytes := wifDecoded.PrivKey.Serialize()
privKeySEC := secp256k1.PrivKeyFromBytes(privKeyBytes)

// Step 2: Sign with Decred's SignCompact (embeds recovery ID automatically)
compactSig := ecdsa.SignCompact(privKeySEC, digest[:], true)

// Step 3: Extract components
recoveryByte := compactSig[0]
rBytes := compactSig[1:33]
sBytes := compactSig[33:65]

// Step 4: Extract recovery ID from recovery byte
recoveryID := int(recoveryByte) - 31

// Step 5: Check and canonicalize s if needed
s := new(big.Int).SetBytes(sBytes)
nDiv2 := new(big.Int)
nDiv2.SetString("7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0", 16)

if s.Cmp(nDiv2) > 0 {
    // Canonicalize: s = N - s
    nValue := new(big.Int)
    nValue.SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
    s = new(big.Int).Sub(nValue, s)

    sBytes = s.Bytes()
    if len(sBytes) < 32 {
        sBytes = append(make([]byte, 32-len(sBytes)), sBytes...)
    }

    // CRITICAL: Flip recovery ID bit 0 when s is flipped
    recoveryID = recoveryID ^ 1
}

// Step 6: Build final signature
canonical := append(rBytes, sBytes...)
finalSig := append([]byte{byte(27 + 4 + recoveryID)}, canonical...)
tx.Signatures = append(tx.Signatures, hex.EncodeToString(finalSig))
```

## Wire Format Conversion

### Location: `types/types.go` - Amount.Bytes() method

```go
// Automatically converts HIVE/HBD to STEEM/SBD for signing
wireSymbol := WireSymbolAliases[a.Symbol]  // HIVE → STEEM
if wireSymbol == "" {
    wireSymbol = a.Symbol  // Use as-is if no alias
}

// Binary format: [amount_int64LE][precision_uint8][symbol_7bytes_padded]
```

## Key Constants

```go
// HIVE Chain ID (used in message digest)
const HIVE_CHAIN_ID = "beeab0de00000000000000000000000000000000000000000000000000000000"

// secp256k1 curve order
const N = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141

// N/2 (canonicalization threshold)
const N_DIV_2 = 0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0

// Wire symbol aliases (Steem fork compatibility)
var WireSymbolAliases = map[string]string{
    "HIVE": "STEEM",
    "HBD":  "SBD",
}
```

## Recovery ID Encoding

```text
recovery_id: 0-3 (2 bits)
  Bit 0: Y-coordinate parity (even=0, odd=1)
  Bit 1: X-coordinate overflow (0=normal, 1=overflowed by N)

Final signature byte: 27 + 4 + recovery_id
  27: Bitcoin's base value
  4:  Compressed public key indicator
  recovery_id: Which of 4 recovery possibilities
  Result: 31-34 range
```

## Transaction Signing Flow

```text
Input Data
  ↓
create tx → add operations → get tx hex from node → hash → sign
  ↓
Extract (r, s, recovery_id)
  ↓
if s > N/2:
  s = N - s
  recovery_id = recovery_id ^ 1
  ↓
Build: [27+4+recovery_id][r][s] → hex encode
  ↓
Broadcast
```

## Common Operations

### Transfer

```go
&transaction.Transfer{
    From:   "sender",
    To:     "receiver",
    Amount: "1.000 HIVE",  // Or "1.500 HBD"
    Memo:   "Payment",
}
```

### Vote

```go
&transaction.Vote{
    Voter:    "account",
    Author:   "post_author",
    Permlink: "post_permlink",
    Weight:   10000,  // 100% = 10000
}
```

### Comment

```go
&transaction.Comment{
    ParentAuthor:   "",
    ParentPermlink: "",
    Author:         "account",
    Permlink:       "post-permlink",
    Title:          "Post Title",
    Body:           "Post body content",
    JSONMetadata:   `{"tags":["tag1"]}`,
}
```

## API Methods

### Client

```go
api := client.NewClient([]string{"https://api.hive.blog"}, 30)
api.Call(apiName, method, params)
api.GetDynamicGlobalProperties()
api.GetAccount(username)
```

### Account

```go
acc := account.NewAccount("username", api)
acc.Refresh()
acc.GetVotingPower()
acc.GetRCInfo()
```

### Transaction

```go
tx := transaction.NewTransaction(api)
tx.AppendOp(operation)
tx.Sign(wifKey)
tx.Broadcast()
```

### Wallet

```go
w := wallet.NewWallet()
w.AddKey(username, keyType, wifKey)
w.Sign(tx, username, keyType)
```

## Error Handling

```go
// Transaction errors
if err := tx.Sign(wif); err != nil {
    log.Fatal(err)  // Handles signing issues
}

if err := w.AddKey(user, "active", wif); err != nil {
    log.Fatal(err)  // Handles invalid WIF
}
```

## Testing

```bash
# Build all packages
go build ./...

# Run account query
go run main.go

# Build transfer example
go build -o examples/transfer ./examples

# Test with WIF
export ACTIVE_WIF="5Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
./examples/transfer
```

## Debugging

### Check Signature Format

```go
sig := append([]byte{byte(31 + recoveryID)}, canonical...)
if len(sig) == 65 {
    fmt.Printf("✓ Signature is 65 bytes\n")
    fmt.Printf("  Recovery: %d (0-3)\n", recoveryID)
    fmt.Printf("  Hex: %s\n", hex.EncodeToString(sig))
}
```

### Verify S Canonicalization

```go
s := new(big.Int).SetBytes(sBytes)
nDiv2 := new(big.Int)
nDiv2.SetString("7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0", 16)

if s.Cmp(nDiv2) > 0 {
    fmt.Println("⚠ S > N/2 - needs canonicalization")
} else {
    fmt.Println("✓ S is canonical")
}
```

### Check Wire Format

```go
amt, _ := types.ParseAmount("1.000 HIVE")
amtBytes, _ := amt.Bytes()
// Should contain "STEEM" in wire format, not "HIVE"
```

## Performance Tips

1. **Reuse Client**: Create once, use for multiple calls
2. **Node Failover**: Provide multiple nodes for resilience
3. **Batch Operations**: Add multiple ops to one transaction
4. **Cache Account Data**: Don't Refresh() on every operation

## Common Mistakes to Avoid

❌ **Not flipping recovery ID when s is canonicalized**

- Result: "unable to reconstruct public key from signature"
- Fix: `recovery_id = recovery_id ^ 1` when `s > N/2`

❌ **Using STEEM/SBD in user input**

- Result: Signatures won't match
- Fix: Always use HIVE/HBD, conversion happens automatically

❌ **Not canonicalizing s**

- Result: "is_canonical(c): signature is not canonical"
- Fix: Check `if s > N/2: s = N - s`

❌ **Using wrong chain ID**

- Result: Signature verification fails
- Fix: Use `HIVE_CHAIN_ID`, not Steem's chain ID

## Resources

- **README.md**: Quick start and API reference
- **IMPLEMENTATION_SUMMARY.md**: Complete feature list
- **SIGNING_IMPLEMENTATION.md**: Detailed signing process
- **RECOVERY_ID_DEEP_DIVE.md**: Recovery ID mathematics
- **WIRE_FORMAT.md**: Wire format explanation
- **CANONICAL_SIGNATURES.md**: Canonicalization details

## Version Info

- **Implementation**: Go 1.16+
- **Status**: Production-Ready ✅
- **Compatibility**: Python anther 1:1 match
- **Blockchain**: HIVE mainnet compatible

---

**Last Updated**: October 18, 2025
**Status**: ✅ COMPLETE AND TESTED
