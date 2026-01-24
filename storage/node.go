// Package storage provides decentralized message storage
package storage

import (
	"context"

	"github.com/parsdao/node/config"
)

// Node is a storage node for encrypted messages
type Node struct {
	cfg     config.StorageConfig
	running bool
}

// NewNode creates a new storage node
func NewNode(cfg config.StorageConfig) (*Node, error) {
	return &Node{
		cfg: cfg,
	}, nil
}

// Start starts the storage node
func (n *Node) Start(ctx context.Context) error {
	n.running = true
	// TODO: Initialize storage backend
	return nil
}

// Stop stops the storage node
func (n *Node) Stop() {
	n.running = false
}

// Store stores an encrypted message
func (n *Node) Store(ctx context.Context, key string, data []byte, ttl int64) error {
	// TODO: Store encrypted data with TTL
	return nil
}

// Retrieve retrieves stored data
func (n *Node) Retrieve(ctx context.Context, key string) ([]byte, error) {
	// TODO: Retrieve encrypted data
	return nil, nil
}

// Delete deletes stored data
func (n *Node) Delete(ctx context.Context, key string) error {
	// TODO: Delete data
	return nil
}
