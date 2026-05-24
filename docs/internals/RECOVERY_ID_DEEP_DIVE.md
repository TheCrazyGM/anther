# Recovery ID and Signature Canonicalization: Deep Dive

## The Problem and Solution

When signing HIVE transactions with Go, we initially got the error:

```text
"unable to reconstruct public key from signature"
```

**Root Cause**: Used wrong cryptographic library (go-ethereum's `crypto.Sign()`) that doesn't support recovery ID testing with canonicalized signatures.

**Solution**: Use Decred's secp256k1 library which provides `SignCompact()` that properly handles recovery ID embedding and canonicalization together.

## Understanding Recovery IDs

### What is a Recovery ID?

In ECDSA, a signature consists of `(r, s)` values. However, there's not a one-to-one mapping from `(r, s)` to a public key. Multiple public keys could potentially produce the same `(r, s)` pair through different elliptic curve math paths.

The **recovery ID** (0-3) tells us which of these paths was taken, allowing us to recover the exact public key from just `(r, s)` and the message hash.

### Recovery ID Encoding

```text
recovery_id: 0-3 (2 bits)
  Bit 0: Y-coordinate parity (0=even, 1=odd)
  Bit 1: X-coordinate overflow (0=normal, 1=overflowed by N)
```

In Bitcoin/HIVE signatures, this is encoded as:

```text
final_byte = 27 + 4 + recovery_id
           = 31-34 (for compressed public keys)
```

Where:

- **27**: Bitcoin's arbitrary base value
- **4**: Flag indicating compressed public key
- **recovery_id**: Which of 4 possibilities

## The S Canonicalization Impact

### Original Signature from ECDSA

```text
crypto.Sign() produces:
sig = [recovery_id+27+4][r: 32 bytes][s: 32 bytes]
```

At this point, `s` might be > N/2.

### Canonicalization Requirement

HIVE requires canonical signatures where `s ≤ N/2`:

```go
if s > N/2 {
    s_canonical = N - s
}
```

### Why This Breaks Recovery

Here's the critical insight: **When you flip s, you're changing the mathematical relationship between the points on the elliptic curve.**

The recovery formula is:

```text
Q = r_inv * (s*R - e*G)
```

If you change `s` to `N - s`:

```text
Q' = r_inv * ((N-s)*R - e*G)
   = r_inv * (N*R - s*R - e*G)
   = r_inv * (-s*R - e*G)   [since N*R = 0 on the curve]
   = -Q
```

The recovered public key is **negated** (y-coordinate flipped). On an elliptic curve, this means:

- If y was even (parity 0), it's now odd (parity 1)
- If y was odd (parity 1), it's now even (parity 0)

Therefore, **we must flip recovery ID bit 0** (the y-parity bit).

## The Correct Solution: Using Decred's secp256k1

### Algorithm Using SignCompact()

```go
// Step 1: Convert WIF to secp256k1 private key
wifDecoded, _ := btcutil.DecodeWIF(wif)
privKeyBytes := wifDecoded.PrivKey.Serialize()
privKeySEC := secp256k1.PrivKeyFromBytes(privKeyBytes)

// Step 2: Sign with SignCompact (embeds recovery ID automatically)
compactSig := ecdsa.SignCompact(privKeySEC, digest[:], true)

// Step 3: Extract components
recoveryByte := compactSig[0]  // Already encoded as 27 + recovery_id + 4
rBytes := compactSig[1:33]
sBytes := compactSig[33:65]

// Step 4: Extract recovery ID
recoveryID := int(recoveryByte) - 31

// Step 5: Check if s needs canonicalization
s := new(big.Int).SetBytes(sBytes)
if s.Cmp(nDiv2) > 0 {
    s = new(big.Int).Sub(N, s)  // Canonicalize: s = N - s
    sBytes = s.Bytes()
    if len(sBytes) < 32 {
        sBytes = append(make([]byte, 32-len(sBytes)), sBytes...)
    }

    // Step 6: Flip recovery ID bit 0 when s is flipped
    recoveryID = recoveryID ^ 1
}

// Step 7: Create final signature
finalSig := append([]byte{byte(27 + 4 + recoveryID)}, append(rBytes, sBytes...)...)
```

### Why This Works

1. **SignCompact() handles complexity**: Returns properly formatted compact signature
2. **Recovery ID embedded**: First byte already contains the recovery information
3. **Canonicalization safe**: The recovery ID adjustment works with canonicalized s
4. **Efficient**: No need to test all 4 recovery IDs

### Why Only Bit 0 Flips?

When we canonicalize `s` to `N - s`, we're affecting the y-coordinate of the elliptic curve point:

```text
The curve equation: y² = x³ + 7 (mod p)
```

For any `x` on the curve, there are two possible `y` values (even and odd parity). When we canonicalize s, the point's y-coordinate gets negated, which flips its parity:

- **Bit 0 (y-parity)**: FLIPPED ← This changes when s is canonicalized
- **Bit 1 (x-overflow)**: UNCHANGED ← The x-coordinate stays the same

Therefore: `recovery_id' = recovery_id ^ 1` (flip only bit 0)

## Python Reference Implementation

The Python cryptography library uses similar logic internally:

```python
# Python anther uses cryptography library's signing
# which produces compact signatures with recovery info
signature_bytes = private_key.sign_compact(message)
# This returns 65-byte compact signature with embedded recovery ID
```

The Python implementation achieves the same result, but through the cryptography library's internal mechanisms.

## Verification Steps

### Test Case: Finding s > N/2

```go
// Try multiple messages until we find one where s > N/2
for attempt := 0; attempt < 100; attempt++ {
    digest := SHA256(message + attempt)
    sig := crypto.Sign(digest, privKey)

    sValue := sig[33:65]  // Extract s bytes

    if sValue > N/2 {
        // Found! Now we can verify our logic
        break
    }
}
```

### Recovery ID Calculation

```text
Original sig:    recovery_byte = 0x56 = 86 decimal
Recovery ID:     (86 - 27) % 4 = 59 % 4 = 3 ✓

After s flip:    recovery_id = 3 ^ 1 = 2 (flipped bit 0)
Final recovery_byte: 27 + 4 + 2 = 33 (0x21)
```

### Final Signature Format

```text
[0x21][r: 32 bytes][canonical_s: 32 bytes] = 65 bytes
```

## Common Mistakes

### ❌ Mistake 1: Not Adjusting Recovery ID

```go
// WRONG: Canonicalize s but keep same recovery ID
if s > N/2 {
    s = N - s
}
// recovery_id stays the same  ← BUG!
```

Result: Signature can't recover the correct public key.

### ❌ Mistake 2: Adjusting Wrong Bit

```go
// WRONG: Flip bit 1 instead of bit 0
recovery_id = recovery_id ^ 2
```

Result: Wrong public key recovered, but not the original.

### ❌ Mistake 3: Flipping All Bits

```go
// WRONG: XOR with 3 (both bits)
recovery_id = recovery_id ^ 3
```

Result: Both y-parity and x-overflow wrong, definitely wrong.

### ✅ Correct: Flip Only Bit 0

```go
// RIGHT: Only flip the y-parity bit
recovery_id = recovery_id ^ 1
```

Result: Signature properly recovers to original public key.

## Mathematical Proof

For a point `P = (x, y)` on the secp256k1 curve:

- Negation gives `−P = (x, −y ≡ p − y mod p)`
- Y-parity flips: even ↔ odd
- X-coordinate unchanged

In ECDSA recovery:

```text
For signature (r, s) with message hash e:
  - recovery_bit_0 indicates y-parity of R point
  - recovery_bit_1 indicates x-overflow of R point

When s → N − s:
  - R point gets negated (−R)
  - Y-parity flips (bit 0 must flip)
  - X-coordinate unchanged (bit 1 stays same)
```

Therefore: `recovery_id' = recovery_id XOR 1`

## References

1. **SEC 2: Recommended Elliptic Curve Domain Parameters**
   - Defines ECDSA signature generation and verification
   - Explains public key recovery

2. **Bitcoin BIP-62: DER Signature Canonicalization**
   - Explains why s must be ≤ N/2
   - Used in Bitcoin and adopted by HIVE

3. **secp256k1 Specification**
   - Elliptic curve parameters used in HIVE
   - Curve equation: y² = x³ + 7 (mod p)

4. **go-ethereum crypto/signature.go**
   - Implementation of Sign() and Ecrecover()
   - Shows how recovery IDs are handled

## Summary

The complete recovery ID adjustment logic:

```text
1. Extract recovery ID from crypto.Sign: (recovery_byte - 27) % 4
2. Check if s > N/2
3. If YES:
   - Canonicalize: s = N - s
   - Flip recovery ID bit 0: recovery_id ^ 1
4. If NO:
   - Keep recovery ID as-is
5. Build final signature: [27 + 4 + adjusted_recovery_id][r][canonical_s]
```

This ensures:

- ✅ Signature is canonical (s ≤ N/2)
- ✅ Public key can be recovered
- ✅ HIVE blockchain accepts the signature
- ✅ Transaction broadcasts successfully

---

**Key Takeaway**: When canonicalizing an ECDSA signature by flipping s, you must also flip the recovery ID's y-parity bit. This maintains the mathematical consistency required for public key recovery.
