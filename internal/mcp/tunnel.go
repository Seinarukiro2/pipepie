package mcp

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/pipepie/pipepie/internal/client"
	"github.com/pipepie/pipepie/internal/config"
)

// tunnelManager manages live tunnel connections spawned by MCP tools.
type tunnelManager struct {
	account *config.Account
	mu      sync.Mutex
	tunnels map[string]*runningTunnel // key: "port" or "subdomain"
}

type runningTunnel struct {
	client    *client.Client
	cancel    context.CancelFunc
	port      string
	subdomain string
	forward   string
}

func newTunnelManager(account *config.Account) *tunnelManager {
	return &tunnelManager{
		account: account,
		tunnels: make(map[string]*runningTunnel),
	}
}

// connect starts a tunnel to the given port with an optional subdomain.
func (tm *tunnelManager) connect(port, subdomain string) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.account == nil || tm.account.Key == "" {
		return "", fmt.Errorf("not logged in — run 'pie login' first")
	}

	// Check if already connected on this port
	if t, ok := tm.tunnels[port]; ok {
		return fmt.Sprintf("already connected: %s → %s", t.subdomain, t.forward), nil
	}

	keyBytes, err := hex.DecodeString(tm.account.Key)
	if err != nil || len(keyBytes) != 32 {
		return "", fmt.Errorf("invalid server key in config")
	}

	forward := "http://localhost:" + port
	cfg := client.Config{
		ServerAddr:   tm.account.Server,
		ServerPubKey: keyBytes,
		Subdomain:    subdomain,
		Forward:      forward,
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := client.New(cfg)

	// Start tunnel in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.Run(ctx)
	}()

	// Brief wait to catch immediate failures
	select {
	case err := <-errCh:
		cancel()
		return "", fmt.Errorf("tunnel failed to start: %w", err)
	default:
	}

	assignedSub := subdomain
	if assignedSub == "" {
		assignedSub = "(auto)"
	}

	tm.tunnels[port] = &runningTunnel{
		client:    c,
		cancel:    cancel,
		port:      port,
		subdomain: assignedSub,
		forward:   forward,
	}

	return fmt.Sprintf("tunnel started: %s → localhost:%s", assignedSub, port), nil
}

// disconnect stops a tunnel by port.
func (tm *tunnelManager) disconnect(port string) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	t, ok := tm.tunnels[port]
	if !ok {
		return "", fmt.Errorf("no tunnel running on port %s", port)
	}

	t.cancel()
	delete(tm.tunnels, port)
	return fmt.Sprintf("disconnected tunnel on port %s (%s)", port, t.subdomain), nil
}

// list returns all active tunnels.
func (tm *tunnelManager) list() []map[string]string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	result := make([]map[string]string, 0, len(tm.tunnels))
	for _, t := range tm.tunnels {
		result = append(result, map[string]string{
			"port":      t.port,
			"subdomain": t.subdomain,
			"forward":   t.forward,
		})
	}
	return result
}
