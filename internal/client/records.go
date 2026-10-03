package client

import (
	"context"
	"encoding/json"
)

// ValidTTL lists the TTL values Njalla accepts.
var ValidTTL = []int{60, 300, 900, 3600, 10800, 21600, 86400}

// ValidPriority lists the priority values Njalla accepts.
var ValidPriority = []int{0, 1, 5, 10, 20, 30, 40, 50, 60}

// Record is a `list-records` entry of any record type.
type Record struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Content      string `json:"content"`
	TTL          int    `json:"ttl"`
	Priority     *int   `json:"prio"`
	Weight       int    `json:"weight"`
	Port         int    `json:"port"`
	Target       string `json:"target"`
	SSHAlgorithm int    `json:"ssh_algorithm"`
	SSHType      int    `json:"ssh_type"`
}

// ListRecords returns every record of the given domain.
func (c *Client) ListRecords(
	ctx context.Context, domain string,
) ([]Record, error) {
	data, err := c.Request(
		ctx, "list-records", map[string]any{"domain": domain},
	)
	if err != nil {
		return nil, err
	}

	var response struct {
		Records []Record `json:"records"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return response.Records, nil
}

// contentParams builds add-record/edit-record params from the fields the
// content-based types use: `name`, `type`, `content`, `ttl` and, if set,
// `id` and `prio`.
func contentParams(domain string, record Record) map[string]any {
	params := map[string]any{
		"domain":  domain,
		"name":    record.Name,
		"type":    record.Type,
		"content": record.Content,
		"ttl":     record.TTL,
	}
	if record.ID != "" {
		params["id"] = record.ID
	}
	if record.Priority != nil {
		params["prio"] = *record.Priority
	}
	return params
}

// AddRecord adds a content-based record (see contentParams) and returns it
// as saved by the API.
func (c *Client) AddRecord(
	ctx context.Context, domain string, record Record,
) (Record, error) {
	data, err := c.Request(ctx, "add-record", contentParams(domain, record))
	if err != nil {
		return Record{}, err
	}

	var saved Record
	if err := json.Unmarshal(data, &saved); err != nil {
		return Record{}, err
	}

	return saved, nil
}

// EditRecord edits a content-based record (see contentParams).
func (c *Client) EditRecord(
	ctx context.Context, domain string, record Record,
) error {
	_, err := c.Request(ctx, "edit-record", contentParams(domain, record))
	return err
}

// RemoveRecord removes the record with the given ID.
func (c *Client) RemoveRecord(
	ctx context.Context, domain string, id string,
) error {
	_, err := c.Request(
		ctx, "remove-record", map[string]any{"domain": domain, "id": id},
	)
	return err
}
