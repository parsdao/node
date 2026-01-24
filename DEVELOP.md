# Local Development

## Prerequisites

Clone the Lux dependencies to `~/work/lux/`:

```bash
mkdir -p ~/work/lux
cd ~/work/lux
git clone https://github.com/luxfi/log
git clone https://github.com/luxfi/crypto
git clone https://github.com/luxfi/sdk
git clone https://github.com/luxfi/evm
git clone https://github.com/luxfi/vm
git clone https://github.com/luxfi/consensus
git clone https://github.com/luxfi/warp
git clone https://github.com/luxfi/ringtail
```

## Using Local Dependencies

Add these replace directives to your `go.mod`:

```go
replace (
	github.com/luxfi/consensus => ~/work/lux/consensus
	github.com/luxfi/crypto => ~/work/lux/crypto
	github.com/luxfi/evm => ~/work/lux/evm
	github.com/luxfi/log => ~/work/lux/log
	github.com/luxfi/ringtail => ~/work/lux/ringtail
	github.com/luxfi/sdk => ~/work/lux/sdk
	github.com/luxfi/vm => ~/work/lux/vm
	github.com/luxfi/warp => ~/work/lux/warp
)
```

Then run:
```bash
go mod tidy
make build
```

## Build Commands

```bash
make build      # Build parsd
make build-gpu  # Build with GPU support
make test       # Run tests
make lint       # Run linter
make devnet     # Run local devnet
```

## Running Validator

### L1 Sovereign Mode
```bash
./bin/parsd --mode=l1 --warp=true --gpu=true
```

### L2 Rollup Mode
```bash
./bin/parsd --mode=l2 --warp=true --gpu=true
```

## Running Alongside Lux Validator

For validators running both Lux mainnet and Pars L2:

```bash
# Terminal 1: Lux mainnet validator
luxd --network-id=mainnet

# Terminal 2: Pars L2 validator (uses Warp to communicate with Lux)
parsd --mode=l2 --warp=true --rpc=127.0.0.1:9660 --p2p=0.0.0.0:9661
```

Both nodes communicate via Warp messaging for cross-chain operations.
