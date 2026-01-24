// Package vm provides virtual machine interfaces and implementations for parsd
package vm

import (
	"context"
)

// VM is the interface for virtual machines running on parsd
type VM interface {
	// Name returns the VM name
	Name() string

	// Start starts the VM
	Start(ctx context.Context) error

	// Stop stops the VM
	Stop() error

	// Health returns the VM health status
	Health() HealthStatus
}

// HealthStatus represents VM health
type HealthStatus struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}
