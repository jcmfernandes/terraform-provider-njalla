package client

import (
	"context"
	"encoding/json"
)

// Server is a `list-servers`/`get-server` entry.
type Server struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	Os          string   `json:"os"`
	Expiry      string   `json:"expiry"`
	Autorenew   bool     `json:"autorenew"`
	SSHKey      string   `json:"ssh_key"`
	Ips         []string `json:"ips"`
	ReverseName string   `json:"reverse_name"`
	OsState     string   `json:"os_state"`
}

// ListServers returns every server of the account.
func (c *Client) ListServers(ctx context.Context) ([]Server, error) {
	data, err := c.Request(ctx, "list-servers", map[string]any{})
	if err != nil {
		return nil, err
	}

	var response struct {
		Servers []Server `json:"servers"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return response.Servers, nil
}
