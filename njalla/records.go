package njalla

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jcmfernandes/terraform-provider-njalla/internal/client"
)

// client.AddRecord/EditRecord always send `content` and `ttl` and nothing
// newer, so these types build requests by hand.

// recordFields lists the attributes each record type sends besides
// `domain`, `type` and `name`, following the per-type rules of add-record.
var recordFields = map[string][]string{
	"ANAME": {"content", "ttl"},
	"DS":    {"content", "ttl"},
	"HTTPS": {"priority", "target"},
	"SRV":   {"content", "ttl", "priority", "weight", "port"},
	"SSHFP": {"content", "ttl", "ssh_algorithm", "ssh_type"},
	"SVCB":  {"priority", "target"},
}

// recordParam maps an attribute to its API parameter name.
func recordParam(attr string) string {
	if attr == "priority" {
		return "prio"
	}
	return attr
}

// recordValue returns the value of an attribute of the given record.
func recordValue(r *client.Record, attr string) any {
	switch attr {
	case "content":
		return r.Content
	case "ttl":
		return r.TTL
	case "priority":
		if r.Priority == nil {
			return 0
		}
		return *r.Priority
	case "weight":
		return r.Weight
	case "port":
		return r.Port
	case "target":
		return r.Target
	case "ssh_algorithm":
		return r.SSHAlgorithm
	case "ssh_type":
		return r.SSHType
	}
	return nil
}

// recordParams builds the add-record/edit-record params for the given type.
// `id` is only included once the record exists.
func recordParams(recordType string, d *schema.ResourceData) map[string]any {
	params := map[string]any{
		"domain": d.Get("domain").(string),
		"type":   recordType,
		"name":   d.Get("name").(string),
	}
	if d.Id() != "" {
		params["id"] = d.Id()
	}
	for _, attr := range recordFields[recordType] {
		params[recordParam(attr)] = d.Get(attr)
	}
	return params
}

func createRecord(
	ctx context.Context, c *client.Client, recordType string,
	d *schema.ResourceData,
) error {
	data, err := c.Request(
		ctx, "add-record", recordParams(recordType, d),
	)
	if err != nil {
		return err
	}

	var saved client.Record
	if err := json.Unmarshal(data, &saved); err != nil {
		return err
	}

	d.SetId(saved.ID)
	return nil
}

func updateRecord(
	ctx context.Context, c *client.Client, recordType string,
	d *schema.ResourceData,
) error {
	_, err := c.Request(
		ctx, "edit-record", recordParams(recordType, d),
	)
	return err
}

// findRecord returns the record with the given ID, or nil if it's gone.
// Entries are decoded one at a time, so a record of another type with an
// unexpected shape can't break the lookup.
func findRecord(
	ctx context.Context, c *client.Client, domain string, id string,
) (*client.Record, error) {
	data, err := c.Request(
		ctx, "list-records", map[string]any{"domain": domain},
	)
	if err != nil {
		return nil, err
	}

	var response struct {
		Records []json.RawMessage `json:"records"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	for _, raw := range response.Records {
		var ref struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(raw, &ref) != nil || ref.ID != id {
			continue
		}

		var found client.Record
		if err := json.Unmarshal(raw, &found); err != nil {
			return nil, err
		}
		return &found, nil
	}

	return nil, nil
}

// setRecord copies the fields of the given type from a record into state.
func setRecord(recordType string, d *schema.ResourceData, r *client.Record) {
	d.Set("name", r.Name)
	for _, attr := range recordFields[recordType] {
		d.Set(attr, recordValue(r, attr))
	}
}
