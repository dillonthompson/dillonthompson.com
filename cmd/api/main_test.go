package main

import "testing"

func TestResolveSiteURL(t *testing.T) {
	cases := []struct{ name, siteURL, siteAddress, want string }{
		{"explicit SITE_URL wins", "https://staging.example.com", "dillonthompson.com", "https://staging.example.com"},
		{"derived from SITE_ADDRESS", "", "www.dillonthompson.com", "https://www.dillonthompson.com"},
		{"local :80 listener falls back", "", ":80", "https://dillonthompson.com"},
		{"nothing set falls back", "", "", "https://dillonthompson.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveSiteURL(tc.siteURL, tc.siteAddress); got != tc.want {
				t.Errorf("resolveSiteURL(%q, %q) = %q, want %q", tc.siteURL, tc.siteAddress, got, tc.want)
			}
		})
	}
}
