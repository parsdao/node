// ParsVM - Post-quantum secure messaging virtual machine
package vm

import (
	"context"
	"fmt"

	"github.com/parsdao/node/config"
	"github.com/parsdao/node/messaging"
	"github.com/parsdao/node/storage"
)

// ParsVM handles post-quantum secure messaging
type ParsVM struct {
	cfg       config.ParsConfig
	storage   *storage.Node
	messenger *messaging.Messenger
	running   bool
}

// NewParsVM creates a new ParsVM instance
func NewParsVM(cfg config.ParsConfig) (*ParsVM, error) {
	if !cfg.Enabled {
		return &ParsVM{cfg: cfg}, nil
	}

	// Initialize storage node
	storageNode, err := storage.NewNode(cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage node: %w", err)
	}

	// Initialize messenger
	messenger, err := messaging.NewMessenger(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create messenger: %w", err)
	}

	return &ParsVM{
		cfg:       cfg,
		storage:   storageNode,
		messenger: messenger,
	}, nil
}

// Name returns the VM name
func (p *ParsVM) Name() string {
	return "pars"
}

// Start starts the ParsVM
func (p *ParsVM) Start(ctx context.Context) error {
	if !p.cfg.Enabled {
		return nil
	}

	// Start storage node
	if p.storage != nil {
		if err := p.storage.Start(ctx); err != nil {
			return fmt.Errorf("failed to start storage: %w", err)
		}
	}

	// Start messenger
	if p.messenger != nil {
		if err := p.messenger.Start(ctx); err != nil {
			return fmt.Errorf("failed to start messenger: %w", err)
		}
	}

	p.running = true
	return nil
}

// Stop stops the ParsVM
func (p *ParsVM) Stop() error {
	p.running = false

	if p.messenger != nil {
		p.messenger.Stop()
	}
	if p.storage != nil {
		p.storage.Stop()
	}

	return nil
}

// Health returns ParsVM health status
func (p *ParsVM) Health() HealthStatus {
	if !p.cfg.Enabled {
		return HealthStatus{Healthy: true, Message: "disabled"}
	}
	if !p.running {
		return HealthStatus{Healthy: false, Message: "not running"}
	}
	return HealthStatus{Healthy: true}
}

// SendMessage sends an encrypted message using PQ crypto
func (p *ParsVM) SendMessage(ctx context.Context, msg *messaging.Message) error {
	if !p.running {
		return fmt.Errorf("ParsVM not running")
	}
	return p.messenger.Send(ctx, msg)
}

// ReceiveMessages retrieves messages for a session
func (p *ParsVM) ReceiveMessages(ctx context.Context, sessionID string) ([]*messaging.Message, error) {
	if !p.running {
		return nil, fmt.Errorf("ParsVM not running")
	}
	return p.messenger.Receive(ctx, sessionID)
}
