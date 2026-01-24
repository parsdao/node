// EVM virtual machine with PQ precompiles
package vm

import (
	"context"
	"fmt"

	"github.com/parsdao/node/config"
)

// EVM wraps the Lux EVM with PQ precompiles
type EVM struct {
	cfg     config.EVMConfig
	running bool
}

// NewEVM creates a new EVM instance
func NewEVM(cfg config.EVMConfig) (*EVM, error) {
	if !cfg.Enabled {
		return &EVM{cfg: cfg}, nil
	}

	return &EVM{
		cfg: cfg,
	}, nil
}

// Name returns the VM name
func (e *EVM) Name() string {
	return "evm"
}

// Start starts the EVM
func (e *EVM) Start(ctx context.Context) error {
	if !e.cfg.Enabled {
		return nil
	}

	// TODO: Initialize EVM with PQ precompiles
	// - ML-DSA at 0x0601
	// - ML-KEM at 0x0603
	// - BLS at 0x0B00
	// - Ringtail at 0x0700
	// - FHE at 0x0800

	e.running = true
	return nil
}

// Stop stops the EVM
func (e *EVM) Stop() error {
	e.running = false
	return nil
}

// Health returns EVM health status
func (e *EVM) Health() HealthStatus {
	if !e.cfg.Enabled {
		return HealthStatus{Healthy: true, Message: "disabled"}
	}
	if !e.running {
		return HealthStatus{Healthy: false, Message: "not running"}
	}
	return HealthStatus{Healthy: true}
}

// Call executes a contract call (placeholder)
func (e *EVM) Call(ctx context.Context, to string, data []byte) ([]byte, error) {
	if !e.running {
		return nil, fmt.Errorf("EVM not running")
	}
	// TODO: Implement actual EVM call
	return nil, nil
}
