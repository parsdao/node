// parsd - Pars Network Node (Sovereign L1)
//
// A sovereign L1 network built on Lux with EVM + SessionVM.
// Post-quantum secure messaging network with PARS native token.
//
// Architecture:
//   - P-Chain: Validator staking in PARS
//   - X-Chain: PARS liquidity and transfers
//   - C-Chain: EVM with PQ precompiles (chain ID 7070)
//   - S-Chain: SessionVM for PQ secure messaging
//
// Usage:
//
//	parsd                     # Run mainnet
//	parsd --testnet           # Run testnet
//	parsd --devnet            # Run local 5-node devnet
//	parsd --network-id=7071   # Custom network

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/luxfi/log"
)

const (
	// Network IDs
	ParsMainnetID = 7070
	ParsTestnetID = 7071
	ParsDevnetID  = 7072

	// VM IDs (base58 encoded)
	EVMID       = "srEXiWaHuhNyGwPUi444Tu47ZEDwxTWrbQiuD7FmgSAQ6X7Dy" // Lux EVM
	SessionVMID = "speKUgLBX6WRD5cfGeEfLa43LxTXUBckvtv4td6F3eTXvRP48" // Session VM

	// Default ports
	DefaultHTTPPort    = 9660
	DefaultStakingPort = 9659
)

var (
	testnet   = flag.Bool("testnet", false, "Run Pars testnet (network-id=7071)")
	devnet    = flag.Bool("devnet", false, "Run Pars devnet (network-id=7072)")
	networkID = flag.Int("network-id", 0, "Network ID (default: 7070 mainnet)")
	httpPort  = flag.Int("http-port", DefaultHTTPPort, "HTTP API port")
	stakingPort = flag.Int("staking-port", DefaultStakingPort, "Staking/P2P port")
	dataDir   = flag.String("data-dir", "", "Data directory (default: ~/.pars)")
	genesis   = flag.String("genesis", "", "Path to genesis file")
	bootstrap = flag.Bool("bootstrap", false, "Bootstrap new network (genesis validators only)")
)

func main() {
	flag.Parse()
	logger := log.New("component", "parsd")

	// Determine network
	netID := ParsMainnetID
	netName := "mainnet"
	if *testnet {
		netID = ParsTestnetID
		netName = "testnet"
	} else if *devnet {
		netID = ParsDevnetID
		netName = "devnet"
	} else if *networkID > 0 {
		netID = *networkID
		netName = "custom"
	}

	// Determine data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("failed to get home directory", "error", err)
		os.Exit(1)
	}

	dataPath := *dataDir
	if dataPath == "" {
		dataPath = filepath.Join(homeDir, ".pars")
	}

	// Ensure directories exist
	pluginDir := filepath.Join(dataPath, "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		logger.Error("failed to create plugin directory", "error", err)
		os.Exit(1)
	}

	// Setup plugins
	if err := setupPlugins(pluginDir, logger); err != nil {
		logger.Error("failed to setup plugins", "error", err)
		os.Exit(1)
	}

	// Build luxd command
	args := buildLuxdArgs(netID, dataPath, pluginDir)

	// Add network-specific flags
	args = append(args,
		fmt.Sprintf("--http-port=%d", *httpPort),
		fmt.Sprintf("--staking-port=%d", *stakingPort),
	)

	// Add genesis if specified or for bootstrap
	if *genesis != "" {
		args = append(args, fmt.Sprintf("--genesis-file=%s", *genesis))
	} else if *bootstrap {
		// Use embedded genesis for bootstrap
		genesisPath := filepath.Join(dataPath, "genesis.json")
		if err := writeEmbeddedGenesis(genesisPath, netName); err != nil {
			logger.Error("failed to write genesis", "error", err)
			os.Exit(1)
		}
		args = append(args, fmt.Sprintf("--genesis-file=%s", genesisPath))
	}

	// Pass through remaining flags
	args = append(args, flag.Args()...)

	logger.Info("starting parsd (Pars Sovereign L1)",
		"network", netName,
		"network-id", netID,
		"datadir", dataPath,
		"plugins", pluginDir,
		"http-port", *httpPort,
		"staking-port", *stakingPort,
	)

	// Find luxd binary
	luxdPath, err := findLuxd()
	if err != nil {
		logger.Error("luxd not found", "error", err)
		logger.Info("Install luxd: go install github.com/luxfi/node/cmd/luxd@latest")
		os.Exit(1)
	}

	// Execute luxd
	cmd := exec.Command(luxdPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	if err := cmd.Start(); err != nil {
		logger.Error("failed to start luxd", "error", err)
		os.Exit(1)
	}

	go func() {
		<-sigCh
		logger.Info("shutting down parsd...")
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			logger.Error("failed to signal luxd", "error", err)
		}
	}()

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		logger.Error("luxd exited with error", "error", err)
		os.Exit(1)
	}
}

// buildLuxdArgs returns the luxd arguments for Pars network
func buildLuxdArgs(networkID int, dataDir, pluginDir string) []string {
	return []string{
		// Network
		fmt.Sprintf("--network-id=%d", networkID),

		// Data directory
		fmt.Sprintf("--data-dir=%s", dataDir),

		// Plugin directory (contains EVM + SessionVM)
		fmt.Sprintf("--plugin-dir=%s", pluginDir),

		// Enable Warp messaging for cross-chain
		"--warp-api-enabled=true",

		// Chain config for PQ precompiles
		"--chain-config-content=" + getParsChainConfig(),

		// Track all chains
		"--track-chains=all",
	}
}

// getParsChainConfig returns the chain configuration with PQ precompiles
func getParsChainConfig() string {
	config := map[string]interface{}{
		"pars-evm": map[string]interface{}{
			// Post-Quantum Cryptography Precompiles
			"precompiles": map[string]string{
				"mldsa":    "0x0601", // ML-DSA-65 signatures
				"mlkem":    "0x0603", // ML-KEM-768 key encapsulation
				"bls":      "0x0B00", // BLS aggregate signatures
				"ringtail": "0x0700", // Ring signatures
				"fhe":      "0x0800", // Fully homomorphic encryption
			},
			// Lux Cross-Chain Precompiles (native access to Lux ecosystem)
			"crossChainPrecompiles": map[string]string{
				"xchain":  "0x1000", // X-Chain: PARS liquidity & staking
				"tchain":  "0x1100", // T-Chain: Trading/DEX access
				"zchain":  "0x1200", // Z-Chain: Zero-knowledge proofs
				"warp":    "0x1300", // Warp: Cross-subnet messaging
				"oracle":  "0x1400", // Oracle: Price feeds
			},
			// DEX/HFT precompiles for native trading
			"dexPrecompiles": map[string]string{
				"lxbook":  "0x2000", // LX orderbook access
				"lxpool":  "0x2100", // LX liquidity pools
				"lxvault": "0x2200", // LX vaults
				"lxfeed":  "0x2300", // LX price feeds (HFT optimized)
			},
		},
		"pars-session": map[string]interface{}{
			"idPrefix":      "07",
			"sessionTTL":    86400,
			"maxMessages":   10000,
			"retentionDays": 30,
		},
		// X-Chain staking configuration
		"pars-staking": map[string]interface{}{
			"minStake":       15000,           // 15,000 PARS minimum
			"lockPeriod":     86400 * 30,      // 30 days lock
			"rewardRate":     0.08,            // 8% APY year 1
			"xchainBridge":   true,            // Enable X-Chain staking bridge
			"feeRecipient":   "X-pars1...",    // X-Chain fee collection
		},
	}
	data, _ := json.Marshal(config)
	return string(data)
}

// writeEmbeddedGenesis writes the network genesis to a file
func writeEmbeddedGenesis(path, network string) error {
	// For now, return an error - in production this would embed the genesis
	// or fetch from a known location
	return fmt.Errorf("embedded genesis not available for %s - use --genesis flag", network)
}

// setupPlugins ensures EVM and SessionVM binaries are in the plugin directory
func setupPlugins(pluginDir string, logger log.Logger) error {
	// Check for EVM plugin
	evmDst := filepath.Join(pluginDir, EVMID)
	if _, err := os.Stat(evmDst); os.IsNotExist(err) {
		evmSrc, err := findEVM()
		if err != nil {
			logger.Warn("EVM plugin not found", "error", err)
		} else {
			if err := os.Symlink(evmSrc, evmDst); err != nil && !os.IsExist(err) {
				return fmt.Errorf("failed to link EVM plugin: %w", err)
			}
			logger.Info("linked EVM plugin", "src", evmSrc, "dst", evmDst)
		}
	}

	// Check for SessionVM plugin
	sessionDst := filepath.Join(pluginDir, SessionVMID)
	if _, err := os.Stat(sessionDst); os.IsNotExist(err) {
		sessionSrc, err := findSessionVM()
		if err != nil {
			logger.Warn("SessionVM plugin not found", "error", err)
		} else {
			if err := os.Symlink(sessionSrc, sessionDst); err != nil && !os.IsExist(err) {
				return fmt.Errorf("failed to link SessionVM plugin: %w", err)
			}
			logger.Info("linked SessionVM plugin", "src", sessionSrc, "dst", sessionDst)
		}
	}

	return nil
}

// findLuxd searches for the luxd binary
func findLuxd() (string, error) {
	if path, err := exec.LookPath("luxd"); err == nil {
		return path, nil
	}

	locations := []string{
		"/usr/local/bin/luxd",
		filepath.Join(os.Getenv("GOPATH"), "bin", "luxd"),
		filepath.Join(os.Getenv("HOME"), "go", "bin", "luxd"),
		filepath.Join(os.Getenv("HOME"), ".lux", "bin", "luxd"),
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", fmt.Errorf("luxd not found in PATH or common locations")
}

// findEVM searches for the EVM plugin binary
func findEVM() (string, error) {
	locations := []string{
		filepath.Join(os.Getenv("HOME"), ".lux", "plugins", EVMID),
		filepath.Join(os.Getenv("HOME"), ".lux", "plugins", "current", EVMID),
		filepath.Join(os.Getenv("GOPATH"), "bin", "evm"),
		"/usr/local/lib/lux/plugins/" + EVMID,
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", fmt.Errorf("EVM plugin not found")
}

// findSessionVM searches for the SessionVM plugin binary
func findSessionVM() (string, error) {
	// Get the directory where parsd binary is located
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	locations := []string{
		// Relative to parsd binary (for development)
		filepath.Join(execDir, "..", "sessionvm", "bin", "sessionvm"),
		filepath.Join(execDir, "..", "..", "sessionvm", "plugin", "sessionvm"),
		// Pars project structure
		filepath.Join(os.Getenv("HOME"), "work", "pars", "sessionvm", "bin", "sessionvm"),
		filepath.Join(os.Getenv("HOME"), "work", "lux", "session", "bin", "sessiond"),
		filepath.Join(os.Getenv("HOME"), "work", "lux", "session", "sessionvm"),
		// Standard plugin locations
		filepath.Join(os.Getenv("HOME"), ".pars", "plugins", SessionVMID),
		filepath.Join(os.Getenv("HOME"), ".lux", "plugins", SessionVMID),
		filepath.Join(os.Getenv("GOPATH"), "bin", "sessionvm"),
		"/usr/local/lib/pars/plugins/" + SessionVMID,
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", fmt.Errorf("SessionVM plugin not found")
}
