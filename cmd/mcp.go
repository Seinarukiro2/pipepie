package cmd

import (
	"github.com/pipepie/pipepie/internal/config"
	piemcp "github.com/pipepie/pipepie/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for Claude, Cursor, and other AI tools",
	Long: `Starts a Model Context Protocol (MCP) server over stdio.

Add to Claude Code:
  claude mcp add --transport stdio pipepie -- pie mcp

Add to Claude Desktop (claude_desktop_config.json):
  {
    "mcpServers": {
      "pipepie": {
        "command": "pie",
        "args": ["mcp"]
      }
    }
  }

The MCP server connects to your pipepie server's API and exposes
tools for listing tunnels, inspecting webhook requests, replaying
requests, and viewing pipeline traces.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.LoadClient()
		active := cfg.ActiveAccount()
		serverHTTP := resolveHTTPAddrFromAccount(cmd, active)
		return piemcp.Run(serverHTTP, active)
	},
}

func init() {
	mcpCmd.Flags().String("server", "", "HTTP API address (default: auto-detect from active account)")
	rootCmd.AddCommand(mcpCmd)
}
