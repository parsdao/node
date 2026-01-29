// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package vm provides the ParsVM implementation using SessionVM from lux/session.
package vm

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/luxfi/ids"
	"github.com/luxfi/log"
	"github.com/luxfi/session/crypto"
	sessionvm "github.com/luxfi/session/vm"
)

// SessionProvider wraps the SessionVM for Pars integration
type SessionProvider struct {
	vm     *sessionvm.VM
	logger log.Logger
}

// NewSessionProvider creates a new SessionProvider
func NewSessionProvider(logger log.Logger) (*SessionProvider, error) {
	factory := &sessionvm.Factory{}
	vm, err := factory.New(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create SessionVM: %w", err)
	}

	return &SessionProvider{
		vm:     vm,
		logger: logger,
	}, nil
}

// Shutdown gracefully stops the SessionVM
func (sp *SessionProvider) Shutdown(ctx context.Context) error {
	return sp.vm.Shutdown(ctx)
}

// GenerateIdentity creates a new PQ identity using ML-KEM-768 and ML-DSA-65
func (sp *SessionProvider) GenerateIdentity() (*crypto.Identity, error) {
	return crypto.GenerateIdentity()
}

// CreateSession creates a new session between participants
func (sp *SessionProvider) CreateSession(ctx context.Context, participantIDs []string, publicKeys [][]byte) (*sessionvm.Session, error) {
	participants := make([]ids.ID, len(participantIDs))
	for i, p := range participantIDs {
		id, err := ids.FromString(p)
		if err != nil {
			return nil, fmt.Errorf("invalid participant ID %s: %w", p, err)
		}
		participants[i] = id
	}

	return sp.vm.CreateSession(participants, publicKeys)
}

// SendMessage sends an encrypted message through a session
func (sp *SessionProvider) SendMessage(ctx context.Context, sessionID, senderID string, ciphertext, signature []byte) (*sessionvm.Message, error) {
	sid, err := ids.FromString(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	sender, err := ids.FromString(senderID)
	if err != nil {
		return nil, fmt.Errorf("invalid sender ID: %w", err)
	}

	return sp.vm.SendMessage(sid, sender, ciphertext, signature)
}

// GetSession retrieves session information
func (sp *SessionProvider) GetSession(ctx context.Context, sessionID string) (*sessionvm.Session, error) {
	sid, err := ids.FromString(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	return sp.vm.GetSession(sid)
}

// CloseSession closes an active session
func (sp *SessionProvider) CloseSession(ctx context.Context, sessionID string) error {
	sid, err := ids.FromString(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	return sp.vm.CloseSession(sid)
}

// Health returns the health status of the SessionVM
func (sp *SessionProvider) Health() HealthStatus {
	result, err := sp.vm.HealthCheck(context.Background())
	if err != nil {
		return HealthStatus{Healthy: false, Message: err.Error()}
	}

	health, ok := result.(map[string]interface{})
	if !ok {
		return HealthStatus{Healthy: false, Message: "invalid health response"}
	}

	healthy, _ := health["healthy"].(bool)
	return HealthStatus{Healthy: healthy}
}

// CreateSecureSession creates a session with full PQ encryption
// 1. Generate identity for local participant
// 2. Create session with remote participant
// 3. Return session ID and local identity
func (sp *SessionProvider) CreateSecureSession(ctx context.Context, localIdentity *crypto.Identity, remoteKEMPublicKey []byte) (*SecureSession, error) {
	// Derive session ID from local and remote public keys
	localKEMPubHex := hex.EncodeToString(localIdentity.KEMPublicKey)
	remoteKEMPubHex := hex.EncodeToString(remoteKEMPublicKey)

	// Create session
	session, err := sp.vm.CreateSession(
		[]ids.ID{}, // Will be populated when we have full participant IDs
		[][]byte{localIdentity.KEMPublicKey, remoteKEMPublicKey},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &SecureSession{
		SessionID:          session.ID.String(),
		LocalIdentity:      localIdentity,
		LocalKEMPublicKey:  localKEMPubHex,
		RemoteKEMPublicKey: remoteKEMPubHex,
		Status:             session.Status,
	}, nil
}

// SecureSession represents a secure messaging session with PQ crypto
type SecureSession struct {
	SessionID          string
	LocalIdentity      *crypto.Identity
	LocalKEMPublicKey  string
	RemoteKEMPublicKey string
	Status             string
}

// EncryptMessage encrypts a message for the remote participant
func (ss *SecureSession) EncryptMessage(plaintext []byte, remoteKEMPublicKey []byte) ([]byte, error) {
	return crypto.EncryptToRecipient(remoteKEMPublicKey, plaintext)
}

// DecryptMessage decrypts a message from the remote participant
func (ss *SecureSession) DecryptMessage(ciphertext []byte) ([]byte, error) {
	return ss.LocalIdentity.DecryptFrom(ciphertext)
}

// SignMessage signs a message with the local identity
func (ss *SecureSession) SignMessage(message []byte) ([]byte, error) {
	return ss.LocalIdentity.SignMessage(message)
}
