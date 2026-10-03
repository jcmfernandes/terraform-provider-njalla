package njalla

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestRecordParamsPerType(t *testing.T) {
	cases := []struct {
		recordType string
		resource   *schema.Resource
		raw        map[string]any
		expected   map[string]any
	}{
		{
			"ANAME", resourceRecordANAME(),
			map[string]any{
				"domain": "testing.com", "name": "@",
				"ttl": 3600, "content": "target.testing.com",
			},
			map[string]any{
				"domain": "testing.com", "type": "ANAME", "name": "@",
				"ttl": 3600, "content": "target.testing.com",
			},
		},
		{
			"DS", resourceRecordDS(),
			map[string]any{
				"domain": "testing.com", "name": "sub",
				"ttl": 3600, "content": "1 13 2 abcdef",
			},
			map[string]any{
				"domain": "testing.com", "type": "DS", "name": "sub",
				"ttl": 3600, "content": "1 13 2 abcdef",
			},
		},
		{
			"HTTPS", resourceRecordHTTPS(),
			map[string]any{
				"domain": "testing.com", "name": "@",
				"priority": 1, "target": ".",
			},
			map[string]any{
				"domain": "testing.com", "type": "HTTPS", "name": "@",
				"prio": 1, "target": ".",
			},
		},
		{
			"SRV", resourceRecordSRV(),
			map[string]any{
				"domain": "testing.com", "name": "_sip._tcp",
				"ttl": 3600, "priority": 10, "weight": 5, "port": 5060,
				"content": "sip.testing.com",
			},
			map[string]any{
				"domain": "testing.com", "type": "SRV", "name": "_sip._tcp",
				"ttl": 3600, "prio": 10, "weight": 5, "port": 5060,
				"content": "sip.testing.com",
			},
		},
		{
			"SSHFP", resourceRecordSSHFP(),
			map[string]any{
				"domain": "testing.com", "name": "host",
				"ttl": 3600, "ssh_algorithm": 4, "ssh_type": 2,
				"content": "abcdef",
			},
			map[string]any{
				"domain": "testing.com", "type": "SSHFP", "name": "host",
				"ttl": 3600, "ssh_algorithm": 4, "ssh_type": 2,
				"content": "abcdef",
			},
		},
		{
			"SVCB", resourceRecordSVCB(),
			map[string]any{
				"domain": "testing.com", "name": "_dns",
				"priority": 1, "target": "dns.testing.com",
			},
			map[string]any{
				"domain": "testing.com", "type": "SVCB", "name": "_dns",
				"prio": 1, "target": "dns.testing.com",
			},
		},
	}

	for _, c := range cases {
		d := schema.TestResourceDataRaw(t, c.resource.Schema, c.raw)

		result := recordParams(c.recordType, d)
		if !reflect.DeepEqual(result, c.expected) {
			t.Errorf("%s: got %v, expected %v", c.recordType, result, c.expected)
		}

		d.SetId("1234")
		c.expected["id"] = "1234"

		result = recordParams(c.recordType, d)
		if !reflect.DeepEqual(result, c.expected) {
			t.Errorf("%s: got %v, expected %v", c.recordType, result, c.expected)
		}
	}
}
