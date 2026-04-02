package mcp

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/pipepie/pipepie/internal/config"
)

// Run starts the MCP stdio server connected to the pipepie API at serverURL.
// It blocks until stdin is closed or the context is cancelled.
func Run(serverURL string, account *config.Account) error {
	client := NewClient(serverURL)
	tm := newTunnelManager(account)

	s := server.NewMCPServer(
		"pipepie",
		"0.1.0",
		server.WithToolCapabilities(true),
	)
	registerTools(s, client, tm)

	stdio := server.NewStdioServer(s)
	return stdio.Listen(context.Background(), os.Stdin, os.Stdout)
}
