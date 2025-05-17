package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExpandSynthetic(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "host-[001:003]",
			expected: []string{"host-001", "host-002", "host-003"},
		},
		{
			input: "rack-[1:2]-node-[01:02]",
			expected: []string{
				"rack-1-node-01",
				"rack-1-node-02",
				"rack-2-node-01",
				"rack-2-node-02",
			},
		},
		{
			input: "mysite-[dca,dcb,dcc]-[01:02]",
			expected: []string{
				"mysite-dca-01",
				"mysite-dca-02",
				"mysite-dcb-01",
				"mysite-dcb-02",
				"mysite-dcc-01",
				"mysite-dcc-02",
			},
		},
		{
			input:    "plainhost",
			expected: []string{"plainhost"},
		},
	}

	for _, tt := range tests {
		got, err := expandSynthetic(tt.input)
		if err != nil {
			t.Errorf("expandSynthetic(%q) returned error: %v", tt.input, err)
			continue
		}
		if diff := cmp.Diff(tt.expected, got); diff != "" {
			t.Errorf("expandSynthetic(%q) mismatch (-expected +got):\n%s", tt.input, diff)
		}
	}
}
