package handlers

import "testing"

func TestEtagMatches(t *testing.T) {
	const etag = `"abc123"`
	cases := []struct {
		name, header string
		want         bool
	}{
		{"empty", "", false},
		{"exact", `"abc123"`, true},
		{"weak (Caddy/Cloudflare compression)", `W/"abc123"`, true},
		{"list with match", `"zzz", W/"abc123"`, true},
		{"list without match", `"zzz", "yyy"`, false},
		{"different tag", `"other"`, false},
		{"wildcard", `*`, true},
		{"partial must not match", `"abc"`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := etagMatches(tc.header, etag); got != tc.want {
				t.Errorf("etagMatches(%q) = %v, want %v", tc.header, got, tc.want)
			}
		})
	}
}
