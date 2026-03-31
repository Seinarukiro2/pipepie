package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ClientConfig is the root config file with multiple accounts.
type ClientConfig struct {
	Active   string              `yaml:"active"`
	Accounts map[string]*Account `yaml:"accounts"`
}

// Account is a single server connection.
type Account struct {
	Type      string            `yaml:"type"`                // "self-hosted" or "managed"
	Server    string            `yaml:"server"`              // e.g. "tunnel.mysite.com:9443"
	Key       string            `yaml:"key"`                 // server public key (hex)
	Subdomain string            `yaml:"subdomain,omitempty"` // default subdomain
	Plan      string            `yaml:"plan,omitempty"`      // managed only: "free", "pro"
	Tunnels   map[string]string `yaml:"tunnels,omitempty"`   // port → subdomain cache
}

// GetTunnelName returns cached subdomain for a port, or empty.
func (a *Account) GetTunnelName(port string) string {
	if a.Tunnels == nil {
		return ""
	}
	return a.Tunnels[port]
}

// SetTunnelName saves subdomain for a port.
func (a *Account) SetTunnelName(port, subdomain string) {
	if a.Tunnels == nil {
		a.Tunnels = make(map[string]string)
	}
	a.Tunnels[port] = subdomain
}

// ClientConfigPath returns ~/.pipepie/config.yaml
func ClientConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pipepie", "config.yaml")
}

// LoadClient reads the config. Returns empty config if not found.
func LoadClient() (*ClientConfig, error) {
	path := ClientConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ClientConfig{Accounts: make(map[string]*Account)}, nil
		}
		return nil, err
	}
	var c ClientConfig
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if c.Accounts == nil {
		c.Accounts = make(map[string]*Account)
	}
	return &c, nil
}

// SaveClient writes the config.
func SaveClient(c *ClientConfig) error {
	path := ClientConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// ActiveAccount returns the currently active account, or nil.
func (c *ClientConfig) ActiveAccount() *Account {
	if c.Active == "" || c.Accounts == nil {
		return nil
	}
	return c.Accounts[c.Active]
}

// AddAccount adds or updates an account and sets it as active.
func (c *ClientConfig) AddAccount(name string, acc *Account) {
	if c.Accounts == nil {
		c.Accounts = make(map[string]*Account)
	}
	c.Accounts[name] = acc
	c.Active = name
}

// RemoveAccount removes an account. If it was active, clears active.
func (c *ClientConfig) RemoveAccount(name string) error {
	if _, ok := c.Accounts[name]; !ok {
		return fmt.Errorf("account %q not found", name)
	}
	delete(c.Accounts, name)
	if c.Active == name {
		c.Active = ""
		// Set first remaining as active
		for k := range c.Accounts {
			c.Active = k
			break
		}
	}
	return nil
}

// SetActive switches the active account.
func (c *ClientConfig) SetActive(name string) error {
	if _, ok := c.Accounts[name]; !ok {
		return fmt.Errorf("account %q not found", name)
	}
	c.Active = name
	return nil
}
