// parsd - Pars Network Node
// Post-quantum secure messaging network built on Lux SDK
//
// Modes:
//   - L1 Sovereign: Independent chain with own validators, Warp bridge to Lux
//   - L2 Rollup: Settles on Lux mainnet, shared security
//
// VMs:
//   - EVM: Smart contracts with PQ precompiles (ML-DSA, ML-KEM)
//   - ParsVM: Post-quantum messaging layer

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/luxfi/log"
	"github.com/parsdao/node/config"
	"github.com/parsdao/node/vm"
)

var (
	configPath = flag.String("config", "", "Path to config file")
	mode       = flag.String("mode", "l1", "Network mode: l1 (sovereign) or l2 (rollup)")
	dataDir    = flag.String("datadir", "~/.pars", "Data directory")
	rpcAddr    = flag.String("rpc", "127.0.0.1:9650", "RPC listen address")
	p2pAddr    = flag.String("p2p", "0.0.0.0:9651", "P2P listen address")
	warpEnable = flag.Bool("warp", true, "Enable Warp cross-chain messaging")
	gpuEnable  = flag.Bool("gpu", true, "Enable GPU acceleration for crypto")
)

func main() {
	flag.Parse()

	logger := log.New("component", "parsd")
	logger.Info("starting parsd",
		"mode", *mode,
		"datadir", *dataDir,
		"warp", *warpEnable,
		"gpu", *gpuEnable,
	)

	// Load configuration
	cfg, err := config.Load(*configPath, &config.Options{
		Mode:       config.Mode(*mode),
		DataDir:    *dataDir,
		RPCAddr:    *rpcAddr,
		P2PAddr:    *p2pAddr,
		WarpEnable: *warpEnable,
		GPUEnable:  *gpuEnable,
	})
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Create node
	node, err := NewNode(cfg, logger)
	if err != nil {
		logger.Error("failed to create node", "error", err)
		os.Exit(1)
	}

	// Start node
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := node.Start(ctx); err != nil {
		logger.Error("failed to start node", "error", err)
		os.Exit(1)
	}

	logger.Info("parsd started",
		"rpc", *rpcAddr,
		"p2p", *p2pAddr,
		"mode", *mode,
	)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down parsd...")
	if err := node.Stop(); err != nil {
		logger.Error("error during shutdown", "error", err)
	}
	logger.Info("parsd stopped")
}

// Node represents a Pars network node
type Node struct {
	cfg    *config.Config
	logger log.Logger
	vms    []vm.VM
}

// NewNode creates a new Pars node
func NewNode(cfg *config.Config, logger log.Logger) (*Node, error) {
	node := &Node{
		cfg:    cfg,
		logger: logger,
	}

	// Initialize VMs
	// 1. EVM with PQ precompiles
	evmVM, err := vm.NewEVM(cfg.EVM)
	if err != nil {
		return nil, err
	}
	node.vms = append(node.vms, evmVM)

	// 2. ParsVM for messaging
	parsVM, err := vm.NewParsVM(cfg.Pars)
	if err != nil {
		return nil, err
	}
	node.vms = append(node.vms, parsVM)

	return node, nil
}

// Start starts the node
func (n *Node) Start(ctx context.Context) error {
	for _, vm := range n.vms {
		if err := vm.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the node
func (n *Node) Stop() error {
	for _, vm := range n.vms {
		if err := vm.Stop(); err != nil {
			n.logger.Error("error stopping VM", "vm", vm.Name(), "error", err)
		}
	}
	return nil
}
