// parsd - Pars Network Node
//
// A Lux node configured for Pars Network with EVM + ParsVM plugins.
// Post-quantum secure messaging network built on Lux.
//
// This is a thin wrapper around luxd that:
//   - Auto-configures plugin directory with EVM and ParsVM
//   - Sets Pars network defaults (chain ID 7070, PQ crypto)
//   - Enables Warp for cross-chain messaging
//
// Usage:
//
//	parsd                     # Run with defaults
//	parsd --network-id=7070   # Custom network
//	parsd --http-port=9660    # Custom ports (run alongside luxd)

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/luxfi/log"
)

const (
	// ParsNetworkID is the default network ID for Pars
	ParsNetworkID = 7070

	// VM IDs (base58 encoded)
	// These are computed from the VM names
	EVMID    = "srEXiWaHuhNyGwPUi444Tu47ZEDwxTWrbQiuD7FmgSAQ6X7Dy"  // Lux EVM
	ParsVMID = "2ZbQaVuXHtT7vfJt8FmWEQKAT4NgtPqWEZHg5m3tUvEiSMnQNt" // Pars VM
)

func main() {
	logger := log.New("component", "parsd")

	// Determine data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("failed to get home directory", "error", err)
		os.Exit(1)
	}
	dataDir := filepath.Join(homeDir, ".pars")

	// Ensure directories exist
	pluginDir := filepath.Join(dataDir, "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		logger.Error("failed to create plugin directory", "error", err)
		os.Exit(1)
	}

	// Setup plugins (symlink or copy EVM and ParsVM binaries)
	if err := setupPlugins(pluginDir, logger); err != nil {
		logger.Error("failed to setup plugins", "error", err)
		os.Exit(1)
	}

	// Build luxd command with Pars defaults
	args := buildLuxdArgs(dataDir, pluginDir)

	// Pass through any additional flags
	args = append(args, os.Args[1:]...)

	logger.Info("starting parsd (luxd wrapper)",
		"datadir", dataDir,
		"plugins", pluginDir,
		"network-id", ParsNetworkID,
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

	// Handle signals - forward to luxd
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	if err := cmd.Start(); err != nil {
		logger.Error("failed to start luxd", "error", err)
		os.Exit(1)
	}

	// Wait for signal or process exit
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

// buildLuxdArgs returns the default luxd arguments for Pars network
func buildLuxdArgs(dataDir, pluginDir string) []string {
	return []string{
		// Network
		fmt.Sprintf("--network-id=%d", ParsNetworkID),

		// Data directory
		fmt.Sprintf("--data-dir=%s", dataDir),

		// Plugin directory (contains EVM + ParsVM)
		fmt.Sprintf("--plugin-dir=%s", pluginDir),

		// Enable Warp messaging
		"--warp-api-enabled=true",

		// Chain config for PQ precompiles
		"--chain-config-content=" + getParsChainConfig(),
	}
}

// getParsChainConfig returns the chain configuration with PQ precompiles
func getParsChainConfig() string {
	return `{
  "pars": {
    "precompiles": {
      "mldsa": "0x0601",
      "mlkem": "0x0603",
      "bls": "0x0B00",
      "ringtail": "0x0700",
      "fhe": "0x0800"
    },
    "session": {
      "idPrefix": "07"
    }
  }
}`
}

// setupPlugins ensures EVM and ParsVM binaries are in the plugin directory
func setupPlugins(pluginDir string, logger log.Logger) error {
	// Check for EVM plugin
	evmDst := filepath.Join(pluginDir, EVMID)
	if _, err := os.Stat(evmDst); os.IsNotExist(err) {
		// Try to find and link EVM binary
		evmSrc, err := findEVM()
		if err != nil {
			logger.Warn("EVM plugin not found, chain creation may fail", "error", err)
		} else {
			if err := os.Symlink(evmSrc, evmDst); err != nil && !os.IsExist(err) {
				return fmt.Errorf("failed to link EVM plugin: %w", err)
			}
			logger.Info("linked EVM plugin", "src", evmSrc, "dst", evmDst)
		}
	}

	// Check for ParsVM plugin
	parsDst := filepath.Join(pluginDir, ParsVMID)
	if _, err := os.Stat(parsDst); os.IsNotExist(err) {
		// Try to find and link ParsVM binary
		parsSrc, err := findParsVM()
		if err != nil {
			logger.Warn("ParsVM plugin not found, will use EVM only", "error", err)
		} else {
			if err := os.Symlink(parsSrc, parsDst); err != nil && !os.IsExist(err) {
				return fmt.Errorf("failed to link ParsVM plugin: %w", err)
			}
			logger.Info("linked ParsVM plugin", "src", parsSrc, "dst", parsDst)
		}
	}

	return nil
}

// findLuxd searches for the luxd binary
func findLuxd() (string, error) {
	// Check PATH first
	if path, err := exec.LookPath("luxd"); err == nil {
		return path, nil
	}

	// Check common locations
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
	// Check common locations
	locations := []string{
		filepath.Join(os.Getenv("HOME"), ".lux", "plugins", EVMID),
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

// findParsVM searches for the ParsVM plugin binary
func findParsVM() (string, error) {
	// Check common locations
	locations := []string{
		filepath.Join(os.Getenv("HOME"), ".pars", "plugins", ParsVMID),
		filepath.Join(os.Getenv("GOPATH"), "bin", "parsvm"),
		"/usr/local/lib/pars/plugins/" + ParsVMID,
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", fmt.Errorf("ParsVM plugin not found")
}
