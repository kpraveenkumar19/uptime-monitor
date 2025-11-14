package monitor

import (
	"context"
	"testing"
)

func TestPing(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		URL string
		Up  bool
	}{
		{"encore.dev", true},
		{"google.com", true},
		{"invalid://scheme", false},
	}

	for _, test := range tests {
		resp, err := Ping(ctx, test.URL)
		if err != nil {
			t.Errorf("url %s: unexpected error: %v", test.URL, err)
		} else if resp.Up != test.Up {
			t.Errorf("url %s: got up=%v, want %v", test.URL, resp.Up, test.Up)
		}
	}
}
