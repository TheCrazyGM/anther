# Canonical ECDSA Signatures for HIVE

## The Issue Resolved ✓

**Error**: `Assert Exception: is_canonical(c): signature is not canonical`

**Cause**: HIVE blockchain requires all ECDSA signatures to be in **canonical form** - a specific deterministic representation of the signature values.

**Solution**: After generating an ECDSA signature, ensure the `s` value is canonical by comparing it to `N/2` (where N is the secp256k1 curve order).

## What is Canonical Form?

An ECDSA signature consists of two values: `r` and `s`. For security and consensus reasons, HIVE (and Bitcoin) require signatures where `s ≤ N/2` to ensure determinism.

### Canonical Formula

```
If s > N/2, then:
    s = N - s
```

This transformation:

- Doesn't change the signature's validity
- Ensures deterministic, non-malleability
- Is required by the blockchain consensus rules

## secp256k1 Constants

```
N (curve order) = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
N/2             = 0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0
```

## Implementation Details

### Signature Format

The Decred `ecdsa.SignCompact()` returns:

- **Byte 0**: Recovery info (27 + recovery_id + 4) = 31 to 34 range
- **Bytes 1-32**: r value (256 bits)
- **Bytes 33-64**: s value (256 bits)
- **Total**: 65 bytes

### Canonicalization Process

```go
// Extract s value from signature
sBytes := compactSig[33:65]
s := new(big.Int).SetBytes(sBytes)

// Check if s > N/2
nDiv2 := new(big.Int)
nDiv2.SetString("7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0", 16)

recoveryID := int(compactSig[0]) - 31

if s.Cmp(nDiv2) > 0 {
    // Make it canonical: s = N - s
    nValue := new(big.Int)
    nValue.SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
    s = new(big.Int).Sub(nValue, s)

    // Reconstruct s as 32 bytes
    sBytes = s.Bytes()
    if len(sBytes) < 32 {
        sBytes = append(make([]byte, 32-len(sBytes)), sBytes...)
    }

    // Important: Adjust recovery ID when s is flipped
    recoveryID = recoveryID ^ 1  // Flip bit 0 (y-parity)

    // Rebuild signature: [27 + 4 + recovery_id][r][canonical_s]
    sig = append([]byte{byte(27 + 4 + recoveryID)}, append(rBytes, sBytes...)...)
} else {
    // Already canonical, use as-is
    sig = compactSig
}
```

## Why This Matters

Without canonical signatures:

- ❌ Blockchain rejects the transaction
- ❌ Error: "is_canonical(c): signature is not canonical"
- ❌ Transfer fails silently

With canonical signatures:

- ✅ Signature is valid per consensus rules
- ✅ Transaction is accepted
- ✅ Transfer succeeds

## Complete Signing Flow

```
1. Create message from transaction data
2. Hash message with SHA256
3. Sign digest with ECDSA private key
4. Check if signature is canonical
5. If not canonical: make it canonical (s = N - s)
6. Return 65-byte canonical signature
7. Include in transaction
8. Broadcast to network ✓
```

## Code Example

```go
// Sign with Decred's secp256k1
privKeySEC := secp256k1.PrivKeyFromBytes(privKeyBytes)
compactSig := ecdsa.SignCompact(privKeySEC, digest[:], true)

// Extract components
recoveryID := int(compactSig[0]) - 31
rBytes := compactSig[1:33]
sBytes := compactSig[33:65]

// Make signature canonical if needed
s := new(big.Int).SetBytes(sBytes)
nDiv2 := new(big.Int)
nDiv2.SetString("7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0", 16)

if s.Cmp(nDiv2) > 0 {
    nValue := new(big.Int)
    nValue.SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
    s = new(big.Int).Sub(nValue, s)

    sBytes = s.Bytes()
    if len(sBytes) < 32 {
        sBytes = append(make([]byte, 32-len(sBytes)), sBytes...)
    }

    // Adjust recovery ID when s is flipped
    recoveryID = recoveryID ^ 1
}

// Build final canonical signature
finalSig := append([]byte{byte(27 + 4 + recoveryID)}, append(rBytes, sBytes...)...)

// Now finalSig is canonical and will be accepted by HIVE
tx.Signatures = append(tx.Signatures, hex.EncodeToString(finalSig))
```

## Testing

The fix has been integrated into the transaction signing flow:

- Automatic canonical check in `Sign()` method
- Applied before adding signature to transaction
- Transparent to user

Test with:

```bash
export ACTIVE_WIF="5Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
./examples/transfer
```

If successful:

- ✅ Signature is canonical
- ✅ Transaction broadcasts
- ✅ Error resolved!

## References

- [Bitcoin Signature Canonicalization](https://github.com/bitcoin/bips/blob/master/bip-0062.md)
- [secp256k1 Curve](https://en.wikipedia.org/wiki/Curve25519)
- [HIVE Blockchain Consensus Rules](https://developers.hive.io/)
- [go-ethereum Crypto Package](https://github.com/ethereum/go-ethereum/tree/master/crypto)

## Summary

HIVE requires canonical ECDSA signatures to ensure consensus and security. By checking if `s > N/2` and transforming `s = N - s` when necessary, we ensure all signatures are in the canonical form required by the blockchain.

This is now automatically handled during transaction signing! ✓
