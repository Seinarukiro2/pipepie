// Package server — auto-detection of AI provider webhooks.
//
// Recognizes payloads from Replicate, fal.ai, RunPod, Modal, BentoML, OpenAI
// and extracts job IDs, status, and timing for automatic pipeline tracing.
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ProviderMatch holds detected provider info from a webhook payload.
type ProviderMatch struct {
	Provider   string // "replicate", "fal", "runpod", "modal", "openai", "bentoml"
	JobID      string // unique job/prediction ID (used as trace correlation)
	Status     string // "starting", "processing", "succeeded", "failed"
	StepName   string // derived step name
	Model      string // model name if available
	DurationMs int64  // processing time if available
}

// DetectProvider analyzes headers and body to identify the AI provider.
// Returns nil if no known provider is detected.
func DetectProvider(headers http.Header, body []byte, path string) *ProviderMatch {
	if len(body) == 0 {
		return nil
	}

	// Try each detector in order
	if m := detectReplicate(headers, body); m != nil {
		return m
	}
	if m := detectFal(headers, body); m != nil {
		return m
	}
	if m := detectRunPod(headers, body); m != nil {
		return m
	}
	if m := detectModal(headers, body); m != nil {
		return m
	}
	if m := detectOpenAI(headers, body); m != nil {
		return m
	}

	// MCP JSON-RPC detection
	if m := detectMCP(headers, body); m != nil {
		return m
	}

	// Path-based fallback for common patterns
	if m := detectByPath(path, body); m != nil {
		return m
	}

	return nil
}

// ── MCP (Model Context Protocol) ─────────────────────────────────────
// JSON-RPC 2.0: {"jsonrpc":"2.0","method":"tools/call","params":{...},"id":1}

func detectMCP(headers http.Header, body []byte) *ProviderMatch {
	var rpc struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		ID      any    `json:"id"`
		Params  struct {
			Name string `json:"name"`
		} `json:"params"`
	}
	if err := json.Unmarshal(body, &rpc); err != nil || rpc.JSONRPC != "2.0" {
		return nil
	}
	if rpc.Method == "" {
		return nil
	}

	stepName := "mcp:" + rpc.Method
	if rpc.Params.Name != "" {
		stepName = "mcp:" + rpc.Method + ":" + rpc.Params.Name
	}

	return &ProviderMatch{
		Provider: "mcp",
		JobID:    fmt.Sprintf("%v", rpc.ID),
		Status:   "received",
		StepName: stepName,
	}
}

// ── Replicate ────────────────────────────────────────────────────────
// Payload: {"id":"xxx","version":"xxx","status":"succeeded","output":[...],"metrics":{"predict_time":8.2}}
// Header: webhook-id, webhook-timestamp, webhook-signature

func detectReplicate(headers http.Header, body []byte) *ProviderMatch {
	// Check Replicate-specific headers
	if headers.Get("webhook-id") == "" && headers.Get("Webhook-Id") == "" {
		// Also check body structure
		var p struct {
			ID      string `json:"id"`
			Version string `json:"version"`
			Status  string `json:"status"`
			Model   string `json:"model"`
			Metrics struct {
				PredictTime float64 `json:"predict_time"`
			} `json:"metrics"`
		}
		if err := json.Unmarshal(body, &p); err != nil || p.ID == "" || p.Version == "" {
			return nil
		}
		model := p.Model
		if model == "" {
			model = p.Version
			if len(model) > 20 {
				model = model[:20]
			}
		}
		return &ProviderMatch{
			Provider:   "replicate",
			JobID:      p.ID,
			Status:     p.Status,
			StepName:   "replicate",
			Model:      model,
			DurationMs: int64(p.Metrics.PredictTime * 1000),
		}
	}

	// Has webhook headers — Replicate standard webhook
	var p struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Version string `json:"version"`
		Model   string `json:"model"`
		Metrics struct {
			PredictTime float64 `json:"predict_time"`
		} `json:"metrics"`
	}
	if err := json.Unmarshal(body, &p); err != nil || p.ID == "" {
		return nil
	}
	return &ProviderMatch{
		Provider:   "replicate",
		JobID:      p.ID,
		Status:     p.Status,
		StepName:   "replicate",
		Model:      p.Model,
		DurationMs: int64(p.Metrics.PredictTime * 1000),
	}
}

// ── fal.ai ───────────────────────────────────────────────────────────
// Payload: {"request_id":"xxx","status":"OK","payload":{...}} or {"request_id":"xxx","status":"IN_QUEUE"}
// Header: x-fal-signature (ED25519)

func detectFal(headers http.Header, body []byte) *ProviderMatch {
	if headers.Get("x-fal-signature") != "" {
		var p struct {
			RequestID string `json:"request_id"`
			Status    string `json:"status"`
		}
		if err := json.Unmarshal(body, &p); err != nil || p.RequestID == "" {
			return nil
		}
		status := strings.ToLower(p.Status)
		if status == "ok" {
			status = "succeeded"
		}
		return &ProviderMatch{
			Provider: "fal",
			JobID:    p.RequestID,
			Status:   status,
			StepName: "fal",
		}
	}

	// No header but check payload shape
	var p struct {
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
		Payload   any    `json:"payload"`
	}
	if err := json.Unmarshal(body, &p); err != nil || p.RequestID == "" {
		return nil
	}
	if p.Status == "" && p.Payload == nil {
		return nil
	}
	status := strings.ToLower(p.Status)
	if status == "ok" {
		status = "succeeded"
	}
	return &ProviderMatch{
		Provider: "fal",
		JobID:    p.RequestID,
		Status:   status,
		StepName: "fal",
	}
}

// ── RunPod ───────────────────────────────────────────────────────────
// Payload: {"id":"xxx-xxx","status":"COMPLETED","output":{...},"executionTime":1234}
// or: {"id":"run_xxx","delayTime":123,"executionTime":456,"status":"COMPLETED"}

func detectRunPod(headers http.Header, body []byte) *ProviderMatch {
	var p struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		ExecutionTime int64  `json:"executionTime"`
		DelayTime     int64  `json:"delayTime"`
	}
	if err := json.Unmarshal(body, &p); err != nil || p.ID == "" {
		return nil
	}
	// RunPod IDs are UUIDs or "run_xxx" format, status is uppercase
	if p.Status != "COMPLETED" && p.Status != "FAILED" && p.Status != "IN_PROGRESS" && p.Status != "IN_QUEUE" && p.Status != "CANCELLED" {
		return nil
	}
	status := strings.ToLower(p.Status)
	if status == "completed" {
		status = "succeeded"
	}
	return &ProviderMatch{
		Provider:   "runpod",
		JobID:      p.ID,
		Status:     status,
		StepName:   "runpod",
		DurationMs: p.ExecutionTime,
	}
}

// ── Modal ────────────────────────────────────────────────────────────
// Payload: {"call_id":"xxx","status":"success","result":{...}}

func detectModal(headers http.Header, body []byte) *ProviderMatch {
	var p struct {
		CallID string `json:"call_id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &p); err != nil || p.CallID == "" {
		return nil
	}
	if p.Status != "success" && p.Status != "failure" && p.Status != "pending" {
		return nil
	}
	status := p.Status
	if status == "success" {
		status = "succeeded"
	} else if status == "failure" {
		status = "failed"
	}
	return &ProviderMatch{
		Provider: "modal",
		JobID:    p.CallID,
		Status:   status,
		StepName: "modal",
	}
}

// ── OpenAI (Batch API callbacks) ─────────────────────────────────────
// Payload: {"id":"batch_xxx","object":"batch","status":"completed",...}

func detectOpenAI(headers http.Header, body []byte) *ProviderMatch {
	var p struct {
		ID     string `json:"id"`
		Object string `json:"object"`
		Status string `json:"status"`
		Model  string `json:"model"`
	}
	if err := json.Unmarshal(body, &p); err != nil || p.ID == "" {
		return nil
	}
	if p.Object != "batch" && !strings.HasPrefix(p.ID, "batch_") {
		return nil
	}
	status := p.Status
	if status == "completed" {
		status = "succeeded"
	}
	return &ProviderMatch{
		Provider: "openai",
		JobID:    p.ID,
		Status:   status,
		StepName: "openai",
		Model:    p.Model,
	}
}

// ── Path-based fallback ──────────────────────────────────────────────

func detectByPath(path string, body []byte) *ProviderMatch {
	lower := strings.ToLower(path)

	// Common AI provider paths
	providers := map[string]string{
		"/replicate": "replicate",
		"/fal":       "fal",
		"/runpod":    "runpod",
		"/modal":     "modal",
		"/openai":    "openai",
		"/anthropic": "anthropic",
		"/huggingface": "huggingface",
		"/stability": "stability",
	}

	for prefix, provider := range providers {
		if strings.HasPrefix(lower, prefix) {
			// Try to extract an ID from the body
			var generic struct {
				ID        string `json:"id"`
				RequestID string `json:"request_id"`
				Status    string `json:"status"`
			}
			json.Unmarshal(body, &generic)
			jobID := generic.ID
			if jobID == "" {
				jobID = generic.RequestID
			}
			status := strings.ToLower(generic.Status)
			if status == "" {
				status = "received"
			}
			return &ProviderMatch{
				Provider: provider,
				JobID:    jobID,
				Status:   status,
				StepName: provider,
			}
		}
	}

	return nil
}
