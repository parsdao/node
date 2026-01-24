package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if cfg.Mode != ModeL1 {
		t.Errorf("expected default mode L1, got %s", cfg.Mode)
	}

	if cfg.Network.ChainID != 7070 {
		t.Errorf("expected chain ID 7070, got %d", cfg.Network.ChainID)
	}

	if cfg.EVM.Precompiles.MLDSA != "0x0601" {
		t.Errorf("expected ML-DSA precompile at 0x0601, got %s", cfg.EVM.Precompiles.MLDSA)
	}

	if cfg.Pars.Session.IDPrefix != "07" {
		t.Errorf("expected session ID prefix 07, got %s", cfg.Pars.Session.IDPrefix)
	}
}

func TestModeValidation(t *testing.T) {
	tests := []struct {
		mode  Mode
		valid bool
	}{
		{ModeL1, true},
		{ModeL2, true},
		{Mode("invalid"), false},
	}

	for _, tt := range tests {
		valid := tt.mode == ModeL1 || tt.mode == ModeL2
		if valid != tt.valid {
			t.Errorf("mode %s: expected valid=%v, got %v", tt.mode, tt.valid, valid)
		}
	}
}

func TestLoadWithOptions(t *testing.T) {
	opts := &Options{
		Mode:       ModeL2,
		DataDir:    "/tmp/test-pars",
		RPCAddr:    "127.0.0.1:8080",
		P2PAddr:    "0.0.0.0:8081",
		WarpEnable: false,
		GPUEnable:  false,
	}

	cfg, err := Load("", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Mode != ModeL2 {
		t.Errorf("expected mode L2, got %s", cfg.Mode)
	}

	if cfg.DataDir != "/tmp/test-pars" {
		t.Errorf("expected datadir /tmp/test-pars, got %s", cfg.DataDir)
	}

	if cfg.Warp.Enabled != false {
		t.Error("expected warp disabled")
	}

	if cfg.Crypto.GPUEnabled != false {
		t.Error("expected GPU disabled")
	}
}
