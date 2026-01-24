// Package config provides configuration for parsd
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Mode defines the network mode
type Mode string

const (
	ModeL1 Mode = "l1" // Sovereign L1 with own validators
	ModeL2 Mode = "l2" // L2 rollup settling on Lux
)

// Options are command-line options
type Options struct {
	Mode       Mode
	DataDir    string
	RPCAddr    string
	P2PAddr    string
	WarpEnable bool
	GPUEnable  bool
}

// Config is the full node configuration
type Config struct {
	// Network mode
	Mode Mode `json:"mode"`

	// Data directory
	DataDir string `json:"dataDir"`

	// Network configuration
	Network NetworkConfig `json:"network"`

	// EVM configuration
	EVM EVMConfig `json:"evm"`

	// Pars messaging configuration
	Pars ParsConfig `json:"pars"`

	// Warp cross-chain configuration
	Warp WarpConfig `json:"warp"`

	// Crypto configuration
	Crypto CryptoConfig `json:"crypto"`

	// Consensus configuration
	Consensus ConsensusConfig `json:"consensus"`
}

// NetworkConfig defines network settings
type NetworkConfig struct {
	RPCAddr   string   `json:"rpcAddr"`
	P2PAddr   string   `json:"p2pAddr"`
	ChainID   uint64   `json:"chainId"`
	NetworkID uint32   `json:"networkId"`
	BootNodes []string `json:"bootNodes"`
}

// EVMConfig defines EVM settings
type EVMConfig struct {
	Enabled     bool   `json:"enabled"`
	ChainID     uint64 `json:"chainId"`
	GasLimit    uint64 `json:"gasLimit"`
	GenesisPath string `json:"genesisPath"`

	// PQ Precompiles
	Precompiles PrecompileConfig `json:"precompiles"`
}

// PrecompileConfig defines PQ precompile addresses
type PrecompileConfig struct {
	MLDSA    string `json:"mldsa"`    // 0x0601 - ML-DSA signatures
	MLKEM    string `json:"mlkem"`    // 0x0603 - ML-KEM key encapsulation
	BLS      string `json:"bls"`      // 0x0B00 - BLS signatures
	Ringtail string `json:"ringtail"` // 0x0700 - PQ threshold signatures
	FHE      string `json:"fhe"`      // 0x0800 - Fully homomorphic encryption
}

// ParsConfig defines Pars messaging settings
type ParsConfig struct {
	Enabled bool `json:"enabled"`

	// Storage node configuration
	Storage StorageConfig `json:"storage"`

	// Onion routing configuration
	Onion OnionConfig `json:"onion"`

	// Session management
	Session SessionConfig `json:"session"`
}

// StorageConfig defines storage node settings
type StorageConfig struct {
	Enabled       bool   `json:"enabled"`
	MaxSize       uint64 `json:"maxSize"` // Max storage in bytes
	RetentionDays int    `json:"retentionDays"`
	DataDir       string `json:"dataDir"`
}

// OnionConfig defines onion routing settings
type OnionConfig struct {
	Enabled  bool `json:"enabled"`
	HopCount int  `json:"hopCount"` // Number of routing hops
}

// SessionConfig defines session management settings
type SessionConfig struct {
	IDPrefix        string `json:"idPrefix"` // "07" for PQ sessions
	KeyRotationDays int    `json:"keyRotationDays"`
}

// WarpConfig defines cross-chain settings
type WarpConfig struct {
	Enabled       bool     `json:"enabled"`
	LuxEndpoint   string   `json:"luxEndpoint"`
	AllowedChains []string `json:"allowedChains"`
}

// CryptoConfig defines cryptographic settings
type CryptoConfig struct {
	GPUEnabled bool `json:"gpuEnabled"` // Metal/CUDA acceleration

	// Signature scheme (ML-DSA-65 for NIST Level 3)
	SignatureScheme string `json:"signatureScheme"`

	// KEM scheme (ML-KEM-768 for NIST Level 3)
	KEMScheme string `json:"kemScheme"`

	// Threshold signatures (Ringtail - Ring-LWE based)
	ThresholdScheme string `json:"thresholdScheme"`
}

// ConsensusConfig defines consensus settings
type ConsensusConfig struct {
	// Quasar consensus configuration
	Engine string `json:"engine"` // "quasar"

	// Block time in milliseconds
	BlockTimeMs uint64 `json:"blockTimeMs"`

	// Validator configuration
	Validators ValidatorConfig `json:"validators"`
}

// ValidatorConfig defines validator settings
type ValidatorConfig struct {
	MinStake uint64 `json:"minStake"`
	MaxCount int    `json:"maxCount"`
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Mode:    ModeL1,
		DataDir: "~/.pars",
		Network: NetworkConfig{
			RPCAddr:   "127.0.0.1:9650",
			P2PAddr:   "0.0.0.0:9651",
			ChainID:   7070, // Pars chain ID
			NetworkID: 7070,
		},
		EVM: EVMConfig{
			Enabled:  true,
			ChainID:  7070,
			GasLimit: 30000000,
			Precompiles: PrecompileConfig{
				MLDSA:    "0x0601",
				MLKEM:    "0x0603",
				BLS:      "0x0B00",
				Ringtail: "0x0700",
				FHE:      "0x0800",
			},
		},
		Pars: ParsConfig{
			Enabled: true,
			Storage: StorageConfig{
				Enabled:       true,
				MaxSize:       10 * 1024 * 1024 * 1024, // 10GB
				RetentionDays: 30,
			},
			Onion: OnionConfig{
				Enabled:  true,
				HopCount: 3,
			},
			Session: SessionConfig{
				IDPrefix:        "07", // PQ session ID prefix
				KeyRotationDays: 90,
			},
		},
		Warp: WarpConfig{
			Enabled:     true,
			LuxEndpoint: "https://api.lux.network",
		},
		Crypto: CryptoConfig{
			GPUEnabled:      true,
			SignatureScheme: "ML-DSA-65",
			KEMScheme:       "ML-KEM-768",
			ThresholdScheme: "Ringtail",
		},
		Consensus: ConsensusConfig{
			Engine:      "quasar",
			BlockTimeMs: 2000,
			Validators: ValidatorConfig{
				MinStake: 1000,
				MaxCount: 100,
			},
		},
	}
}

// Load loads configuration from file and applies options
func Load(path string, opts *Options) (*Config, error) {
	cfg := Default()

	// Load from file if provided
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Apply command-line options
	if opts != nil {
		if opts.Mode != "" {
			cfg.Mode = opts.Mode
		}
		if opts.DataDir != "" {
			cfg.DataDir = opts.DataDir
		}
		if opts.RPCAddr != "" {
			cfg.Network.RPCAddr = opts.RPCAddr
		}
		if opts.P2PAddr != "" {
			cfg.Network.P2PAddr = opts.P2PAddr
		}
		cfg.Warp.Enabled = opts.WarpEnable
		cfg.Crypto.GPUEnabled = opts.GPUEnable
	}

	// Expand paths
	cfg.DataDir = expandPath(cfg.DataDir)
	cfg.Pars.Storage.DataDir = filepath.Join(cfg.DataDir, "storage")

	return cfg, nil
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}
