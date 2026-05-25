# Comparison: Anther vs. HiveGo (vsc-eco/hivego)

This document provides a brutally honest, side-by-side comparison of **Anther** and **HiveGo** (`vsc-eco/hivego`, including its parent forks starting from `DeathwingTheBoss/hivego`).

---

## At a Glance

| Feature | Anther | HiveGo | Winner |
| :--- | :--- | :--- | :--- |
| **Package Architecture** | Modular, sub-packages (`wallet`, `crypto`, `client`, `transaction`) | Monolithic / Flat root directory (`package hivego`) | **Anther** |
| **JSON-RPC Performance** | Single request with exponential backoff & failover | Batch request support & rolling average failover stats | **HiveGo** |
| **Offline Wire Serialization**| Robust, zero-dependency `Bytes()` encoding for all major ops | Basic custom serialization, limited op type support | **Anther** |
| **Key Signature Crypto** | Modern Decrec `secp256k1/v4` | Legacy `secp256k1/v2` | **Anther** |
| **Multi-Signature Verification** | Public key recovery & local authority threshold calculations | Single-key signature support only | **Anther** |
| **Private Memo Security** | ECIES encryption/decryption built-in (ECDH + AES + SHA) | Completely missing | **Anther** |
| **Blockchain Streaming** | Concurrent channel streams with consensus lag compensation | Simple, blocking block-by-block fetch loop | **Anther** |
| **External Dependencies** | Clean runtime (only uses standard lib + crypto) | Pulls in custom third-party JSON-RPC client | **Anther** |
| **Date & Time DX** | Auto-unmarshals naive Hive timestamps into native time | Auto-unmarshals naive Hive timestamps into native time | **Tie** |

---

## Brutal Breakdown: Strengths & Weaknesses

### Anther

#### 🟢 The Good (Strengths)
1. **Unrivaled Crypto Suite**: The inclusion of ECIES private memo encryption and decryption makes it the only Go client capable of handling secure private messaging/transfers out of the box.
2. **Offline-First Security**: Fully encodes transactions and operations locally to bytes, enabling 100% offline signing and verification without trusting any public RPC node.
3. **Advanced Authority Support**: Able to recover public keys from compact signatures and locally check if a transaction meets multi-signature threshold rules before broadcasting.
4. **Resilient Streaming Engine**: Concurrent channel-based block and operation streaming. It backfills missed blocks and handles consensus latency gracefully rather than throwing raw errors on slow blocks.
5. **Clean Dependency Footprint**: Leverages modern decred `secp256k1/v4` and relies entirely on Go's standard library for networking, keeping the runtime codebase secure and auditable.

#### 🔴 The Bad (Weaknesses)
1. **Lack of JSON-RPC Batching**: Currently cannot group multiple distinct JSON-RPC calls into a single HTTP post request. Querying multiple resources requires sequential HTTP roundtrips.
2. **Heavier Build Tree**: Imports `github.com/btcsuite/btcd`, which pulls in a larger tree of Bitcoin utility packages during building, even though we only use it for base58 formatting and basic transaction structures.
3. **High-Level API Completeness**: Missing some quick high-level wrappers for secondary APIs like database parameters, Account History pagination, and Resource Credit mathematical models (these are planned for the upcoming v1.0.0 roadmap).

---

### HiveGo (`vsc-eco/hivego`)

#### 🟢 The Good (Strengths)
1. **JSON-RPC Batching**: Support for batch operations (`rpcExecBatchFast`) allows developers to query ranges of data (like block lists or multiple accounts) in a single network call.
2. **Minimal Code footprint**: Because it implements very few features, it is a lightweight codebase that is quick to read if you just want to do basic node queries.
3. **Rolling Average Failover**: Tracks node success percentages and rolling averages to select the most reliable RPC endpoints.

#### 🔴 The Bad (Weaknesses)
1. **Crude Package Design**: Lacks modular namespace structures. Every Go file is placed directly in the repository root (`package hivego`), mixing client logic, serialization, cryptography, types, and logging into a single namespace.
2. **Outdated Cryptography**: Relies on `secp256k1/v2`, which is outdated and slower compared to modern Go secp256k1 implementations.
3. **No Private Memos**: Lacks any memo encryption/decryption capabilities. Applications cannot read or write encrypted transfers.
4. **Poor Offline Operation Support**: Serializer coverage is very limited, forcing developers to rely on node APIs for complex transaction signing.
5. **No Multi-Signature Tools**: Cannot recover public keys from signatures or verify threshold weights locally.
6. **Brittle Streaming**: The streaming implementation is a basic loop that blocks and can easily crash/panic if a node responds with invalid data or times out.
7. **External Networking Dependencies**: Uses `github.com/cfoxon/jsonrpc2client`, which increases vulnerability surface area and means you don't control the HTTP connection pool directly.
8. **Negligible Test Quality**: The tests are mostly mocks that verify trivial behavior, lacking comprehensive validation of real network payloads and cryptographic operations.

---

## Verdict

* Use **HiveGo** only if you need a tiny, single-purpose client for simple queries, and you heavily rely on JSON-RPC batching.
* Use **Anther** if you are building robust, production-ready applications, bots, or wallets that require reliable streaming, private memo security, secure offline transaction signing, multi-signature configurations, and modular architecture.
