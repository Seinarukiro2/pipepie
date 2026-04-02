package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a lightweight HTTP client for the pipepie server REST API.
type Client struct {
	base   string
	client *http.Client
}

// NewClient creates a new API client with the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		base:   strings.TrimRight(baseURL, "/"),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) get(path string) (json.RawMessage, error) {
	resp, err := c.client.Get(c.base + path)
	if err != nil {
		return nil, fmt.Errorf("cannot reach server at %s: %w", c.base, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}
	return json.RawMessage(body), nil
}

func (c *Client) post(path string, payload any) (json.RawMessage, error) {
	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(data))
	}
	resp, err := c.client.Post(c.base+path, "application/json", bodyReader)
	if err != nil {
		return nil, fmt.Errorf("cannot reach server at %s: %w", c.base, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}
	return json.RawMessage(body), nil
}

func (c *Client) delete(path string) (json.RawMessage, error) {
	req, err := http.NewRequest(http.MethodDelete, c.base+path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot reach server at %s: %w", c.base, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}
	return json.RawMessage(body), nil
}
