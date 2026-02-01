# parsd - Pars Network Node

Post-quantum secure messaging network built on Lux SDK.

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│                          parsd                             │
│                   (Pars Network Node)                      │
├────────────────────────────────────────────────────────────┤
│                    Virtual Machines                        │
│  ┌─────────────────────┐    ┌─────────────────────┐        │
│  │         EVM         │    │      SessionVM      │        │
│  │  (Smart Contracts)  │    │    (PQ Messaging)   │        │
│  ├─────────────────────┤    ├─────────────────────┤        │
│  │ PQ Precompiles:     │    │ • ML-KEM key exch   │        │
│  │ • ML-DSA (0x0601)   │    │ • ML-DSA signatures │        │
│  │ • ML-KEM (0x0603)   │    │ • Onion routing     │        │
│  │ • BLS (0x0B00)      │    │ • Storage nodes     │        │
│  │ • Ringtail (0x0700) │    │ • Session mgmt      │        │
│  │ • FHE (0x0800)      │    │                     │        │
│  └─────────────────────┘    └─────────────────────┘        │
├────────────────────────────────────────────────────────────┤
│                      Warp Bridge                           │
│              (Cross-chain with Lux mainnet)                │
├────────────────────────────────────────────────────────────┤
│                    Quasar Consensus                        │
│           (BLS + Ringtail PQ threshold finality)           │
├────────────────────────────────────────────────────────────┤
│                     lux/crypto                             │
│        (ML-DSA, ML-KEM, Ringtail, BLS - GPU accel)         │
└────────────────────────────────────────────────────────────┘
```

## Modes

### L1 Sovereign
Independent chain with own validators. Uses Warp to bridge to Lux mainnet.

```bash
parsd --mode=l1 --warp=true
```

### L2 Rollup
Settles on Lux mainnet with shared security.

```bash
parsd --mode=l2 --warp=true
```

## Quick Start

```bash
# Build
make build

# Run local devnet
make devnet

# Run L1 sovereign
make run

# Run L2 rollup
make run-l2
```

## Configuration

Default config location: `~/.pars/config.json`

```json
{
  "mode": "l1",
  "network": {
    "chainId": 7070,
    "rpcAddr": "127.0.0.1:9650",
    "p2pAddr": "0.0.0.0:9651"
  },
  "evm": {
    "enabled": true,
    "precompiles": {
      "mldsa": "0x0601",
      "mlkem": "0x0603",
      "bls": "0x0B00",
      "ringtail": "0x0700",
      "fhe": "0x0800"
    }
  },
  "pars": {
    "enabled": true,
    "storage": {
      "maxSize": 10737418240,
      "retentionDays": 30
    },
    "onion": {
      "hopCount": 3
    },
    "session": {
      "idPrefix": "07"
    }
  },
  "warp": {
    "enabled": true,
    "luxEndpoint": "https://api.lux.network"
  },
  "crypto": {
    "gpuEnabled": true,
    "signatureScheme": "ML-DSA-65",
    "kemScheme": "ML-KEM-768",
    "thresholdScheme": "Ringtail"
  },
  "consensus": {
    "engine": "quasar",
    "blockTimeMs": 2000
  }
}
```

## Crypto Stack

All crypto from `lux/crypto` - one implementation, all platforms:

| Algorithm | Purpose | NIST Level |
|-----------|---------|------------|
| ML-DSA-65 | Signatures | Level 3 |
| ML-KEM-768 | Key Encapsulation | Level 3 |
| Ringtail | PQ Threshold Sigs | Ring-LWE |
| BLS12-381 | Consensus | - |
| XChaCha20-Poly1305 | AEAD | - |
| Blake2b/Blake3 | Hashing | - |

GPU acceleration via Metal (macOS/iOS) and CUDA (Linux).

## Session IDs

Pars uses "07" prefix for post-quantum session IDs:

```
07 + hex(Blake2b-256(ML-KEM-768_pk || ML-DSA-65_pk))
```

Legacy Session IDs use "05" prefix (X25519/Ed25519).

## Repository

- GitHub: https://github.com/parsdao/node
- Issues: https://github.com/parsdao/node/issues

## License

MIT
