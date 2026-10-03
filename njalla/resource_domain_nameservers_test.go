package njalla

import (
	"reflect"
	"testing"
)

func TestCustomNameservers(t *testing.T) {
	cases := []struct {
		reported []string
		want     []string
	}{
		{nil, []string{}},
		{
			[]string{"1-you.njalla.no", "2-can.njalla.in", "3-get.njalla.fo."},
			[]string{},
		},
		{[]string{"ns1.example.com"}, []string{"ns1.example.com"}},
		{
			[]string{"1-you.njalla.no", "ns1.example.com"},
			[]string{"1-you.njalla.no", "ns1.example.com"},
		},
	}

	for _, c := range cases {
		if got := customNameservers(c.reported); !reflect.DeepEqual(got, c.want) {
			t.Errorf("customNameservers(%v) = %v, want %v", c.reported, got, c.want)
		}
	}
}
