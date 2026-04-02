package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTools(s *server.MCPServer, c *Client, tm *tunnelManager) {
	s.AddTool(toolOverview, handleOverview(c))
	s.AddTool(toolListTunnels, handleListTunnels(c))
	s.AddTool(toolTunnelStatus, handleTunnelStatus(c))
	s.AddTool(toolListRequests, handleListRequests(c))
	s.AddTool(toolInspectRequest, handleInspectRequest(c))
	s.AddTool(toolReplayRequest, handleReplayRequest(c))
	s.AddTool(toolPipelineTraces, handlePipelineTraces(c))
	s.AddTool(toolTraceTimeline, handleTraceTimeline(c))
	s.AddTool(toolCreateTunnel, handleCreateTunnel(c))
	s.AddTool(toolDeleteTunnel, handleDeleteTunnel(c))
	s.AddTool(toolConnect, handleConnect(tm))
	s.AddTool(toolDisconnect, handleDisconnect(tm))
	s.AddTool(toolActiveTunnels, handleActiveTunnels(tm))
}

// ── Tool definitions ────────────────────────────────────────────────

var toolOverview = mcp.NewTool("overview",
	mcp.WithDescription("Get a dashboard overview of all tunnels with aggregate statistics: total requests, success/error counts, success rates, online status, uptime, and last request time. Best starting point for understanding the current state."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
)

var toolListTunnels = mcp.NewTool("list_tunnels",
	mcp.WithDescription("List all registered webhook tunnels with their online/offline status, protocol, and creation time. Use this to discover available subdomains before inspecting requests."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
)

var toolTunnelStatus = mcp.NewTool("tunnel_status",
	mcp.WithDescription("Check whether a specific tunnel is currently online and what protocol it uses."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithString("subdomain", mcp.Required(), mcp.Description("The tunnel subdomain to check")),
)

var toolListRequests = mcp.NewTool("list_requests",
	mcp.WithDescription("List recent webhook requests received by a tunnel, ordered newest first. Includes method, path, status code, duration, and pipeline metadata. Supports pagination."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithString("subdomain", mcp.Required(), mcp.Description("The tunnel subdomain")),
	mcp.WithNumber("limit", mcp.Description("Number of requests to return, 1-100 (default 20)")),
	mcp.WithNumber("offset", mcp.Description("Pagination offset (default 0)")),
)

var toolInspectRequest = mcp.NewTool("inspect_request",
	mcp.WithDescription("Get full details of a specific webhook request: headers, request body, response status, response body, duration, source IP, and pipeline/trace metadata. Use request IDs from list_requests."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithString("request_id", mcp.Required(), mcp.Description("The request UUID")),
)

var toolReplayRequest = mcp.NewTool("replay_request",
	mcp.WithDescription("Replay a previously captured webhook request to the currently connected tunnel client. The tunnel must be online. Returns the new response status and duration."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithString("subdomain", mcp.Required(), mcp.Description("The tunnel subdomain (must be online)")),
	mcp.WithString("request_id", mcp.Required(), mcp.Description("The request ID to replay")),
)

var toolPipelineTraces = mcp.NewTool("pipeline_traces",
	mcp.WithDescription("List execution traces for a specific AI pipeline (Replicate, fal.ai, RunPod, OpenAI, etc). Pipelines are auto-detected from webhook payloads."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithString("pipeline_id", mcp.Required(), mcp.Description("The pipeline identifier (e.g. 'replicate', 'fal-ai', or custom)")),
	mcp.WithNumber("limit", mcp.Description("Number of traces to return (default 20)")),
)

var toolTraceTimeline = mcp.NewTool("trace_timeline",
	mcp.WithDescription("Get the step-by-step execution timeline for a specific pipeline trace. Shows each step's name, status, duration, and ordering."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithString("trace_id", mcp.Required(), mcp.Description("The trace identifier")),
)

var toolCreateTunnel = mcp.NewTool("create_tunnel",
	mcp.WithDescription("Register a new tunnel subdomain on the server. The subdomain must be lowercase alphanumeric with hyphens, 1-63 characters. Returns the tunnel ID and public URL."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithString("subdomain", mcp.Required(), mcp.Description("Subdomain to register (lowercase alphanumeric and hyphens)")),
)

var toolDeleteTunnel = mcp.NewTool("delete_tunnel",
	mcp.WithDescription("Permanently delete a tunnel and all its stored requests. Cannot be undone."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(true),
	mcp.WithString("subdomain", mcp.Required(), mcp.Description("The tunnel subdomain to delete")),
)

// ── Handlers ────────────────────────────────────────────────────────

func handleOverview(c *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.get("/api/overview")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func handleListTunnels(c *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := c.get("/api/admin/tunnels")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func handleTunnelStatus(c *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sub, err := req.RequireString("subdomain")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.get("/api/tunnels/" + sub + "/status")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func handleListRequests(c *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sub, err := req.RequireString("subdomain")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		limit := req.GetInt("limit", 20)
		offset := req.GetInt("offset", 0)
		path := fmt.Sprintf("/api/tunnels/%s/requests?limit=%d&offset=%d", sub, limit, offset)
		data, err := c.get(path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func handleInspectRequest(c *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("request_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.get("/api/requests/" + id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func handleReplayRequest(c *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sub, err := req.RequireString("subdomain")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := req.RequireString("request_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path := fmt.Sprintf("/api/tunnels/%s/requests/%s/replay", sub, id)
		data, err := c.post(path, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func handlePipelineTraces(c *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pid, err := req.RequireString("pipeline_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		limit := req.GetInt("limit", 20)
		path := fmt.Sprintf("/api/pipelines/%s/traces?limit=%d", pid, limit)
		data, err := c.get(path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func handleTraceTimeline(c *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tid, err := req.RequireString("trace_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.get("/api/traces/" + tid)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func handleCreateTunnel(c *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sub, err := req.RequireString("subdomain")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sub = strings.ToLower(strings.TrimSpace(sub))
		data, err := c.post("/api/admin/tunnels", map[string]string{"subdomain": sub})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func handleDeleteTunnel(c *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sub, err := req.RequireString("subdomain")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := c.delete("/api/admin/tunnels/" + sub)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

// ── Tunnel connection tools ─────────────────────────────────────────

var toolConnect = mcp.NewTool("connect",
	mcp.WithDescription("Start a tunnel that forwards a public URL to a local port. Uses the active account from 'pie login'. Example: connect port 3000 with subdomain 'my-app' to get https://my-app.yourdomain.com → localhost:3000."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithString("port", mcp.Required(), mcp.Description("Local port number to forward to (e.g. '3000', '8080')")),
	mcp.WithString("subdomain", mcp.Description("Subdomain name for the public URL (empty = auto-assigned)")),
)

var toolDisconnect = mcp.NewTool("disconnect",
	mcp.WithDescription("Stop a running tunnel by its local port number."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithString("port", mcp.Required(), mcp.Description("Local port number of the tunnel to stop")),
)

var toolActiveTunnels = mcp.NewTool("active_tunnels",
	mcp.WithDescription("List all tunnels currently connected through this MCP session, showing port, subdomain, and forward address."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
)

func handleConnect(tm *tunnelManager) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		port, err := req.RequireString("port")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		subdomain := req.GetString("subdomain", "")
		msg, err := tm.connect(port, subdomain)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(msg), nil
	}
}

func handleDisconnect(tm *tunnelManager) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		port, err := req.RequireString("port")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		msg, err := tm.disconnect(port)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(msg), nil
	}
}

func handleActiveTunnels(tm *tunnelManager) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tunnels := tm.list()
		if len(tunnels) == 0 {
			return mcp.NewToolResultText("no active tunnels"), nil
		}
		data, _ := json.Marshal(tunnels)
		return mcp.NewToolResultText(string(data)), nil
	}
}
