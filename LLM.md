# Pars Node - AI Assistant Context

## Project Overview

parsd is the Pars Network node - a post-quantum secure messaging network built on Lux SDK.

## Repository Structure

```
~/work/pars/node/
├── cmd/parsd/         # Main entry point
├── config/            # Configuration
├── vm/                # Virtual machines
│   ├── vm.go          # VM interface
│   ├── evm.go         # EVM with PQ precompiles
│   └── pars.go        # ParsVM messaging
├── messaging/         # PQ encrypted messaging
├── storage/           # Decentralized storage
├── go.mod             # github.com/parsdao/node
└── Makefile
```

## Key Concepts

### Modes
- **L1 Sovereign**: Independent chain, own validators, Warp bridge to Lux
- **L2 Rollup**: Settles on Lux mainnet, shared security

### Virtual Machines
1. **EVM**: Smart contracts with PQ precompiles
   - ML-DSA at 0x0601
   - ML-KEM at 0x0603
   - BLS at 0x0B00
   - Ringtail at 0x0700
   - FHE at 0x0800

2. **ParsVM**: Post-quantum messaging
   - ML-KEM-768 key exchange
   - ML-DSA-65 signatures
   - Onion routing
   - Storage nodes
   - Session management

### Crypto
All crypto from `lux/crypto` - single source of truth:
- ML-DSA-65 (FIPS 204) - signatures
- ML-KEM-768 (FIPS 203) - key encapsulation
- Ringtail - PQ threshold signatures (Ring-LWE)
- BLS12-381 - consensus
- XChaCha20-Poly1305 - AEAD
- Blake2b/Blake3 - hashing

GPU acceleration via Metal (macOS/iOS).

### Session IDs
- "07" prefix = post-quantum (ML-KEM + ML-DSA)
- "05" prefix = legacy (X25519 + Ed25519)

Format: `07 + hex(Blake2b-256(KEM_pk || DSA_pk))`

## Dependencies

- github.com/luxfi/sdk - Lux SDK for node building
- github.com/luxfi/crypto - PQ crypto (ML-DSA, ML-KEM)
- github.com/luxfi/ringtail - PQ threshold signatures
- github.com/luxfi/consensus - Quasar consensus
- github.com/luxfi/warp - Cross-chain messaging
- github.com/luxfi/evm - EVM with precompiles

## Commands

```bash
make build      # Build parsd
make run        # Run L1 sovereign
make run-l2     # Run L2 rollup
make devnet     # Run local devnet
make test       # Run tests
```

## Configuration

Default: `~/.pars/config.json`

Key settings:
- `mode`: "l1" or "l2"
- `network.chainId`: 7070
- `crypto.gpuEnabled`: true
- `warp.enabled`: true

## Related Projects

- ~/work/lux/crypto - Go crypto implementations
- ~/work/luxcpp/crypto - C++ GPU (Metal)
- ~/work/lux/accel-rs - Rust FFI bindings
- ~/work/lux/ringtail - PQ threshold signatures
- ~/work/lux/sdk - Lux SDK
- ~/work/session/ - Session fork (Pars clients)
