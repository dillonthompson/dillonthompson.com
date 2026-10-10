package handlers

import (
	"strings"
	"testing"
)

func TestPostInputValidate(t *testing.T) {
	valid := func() postInput {
		return postInput{Slug: "my-post", Title: "My Post", Description: "d", BodyMD: "body", Tags: []string{"go"}, Status: "draft"}
	}
	cases := []struct {
		name    string
		mutate  func(*postInput)
		wantErr string // substring; "" means valid
	}{
		{"valid", func(*postInput) {}, ""},
		{"empty status defaults to draft", func(in *postInput) { in.Status = "" }, ""},
		{"missing title", func(in *postInput) { in.Title = "  " }, "title is required"},
		{"long title", func(in *postInput) { in.Title = strings.Repeat("a", 201) }, "title is too long"},
		{"uppercase slug", func(in *postInput) { in.Slug = "My-Post" }, "slug"},
		{"slug with slash", func(in *postInput) { in.Slug = "a/b" }, "slug"},
		{"slug path traversal", func(in *postInput) { in.Slug = "../etc" }, "slug"},
		{"slug double hyphen", func(in *postInput) { in.Slug = "a--b" }, "slug"},
		{"empty slug", func(in *postInput) { in.Slug = "" }, "slug"},
		{"long slug", func(in *postInput) { in.Slug = strings.Repeat("a", 101) }, "slug"},
		{"long description", func(in *postInput) { in.Description = strings.Repeat("a", 301) }, "description is too long"},
		{"huge body", func(in *postInput) { in.BodyMD = strings.Repeat("a", maxBodyBytes+1) }, "body is too long"},
		{"NUL byte in body", func(in *postInput) { in.BodyMD = "a\x00b" }, "invalid character"},
		{"bad status", func(in *postInput) { in.Status = "archived" }, "status"},
		{"publish without body", func(in *postInput) { in.Status = "published"; in.BodyMD = " " }, "needs a body"},
		{"publish with body", func(in *postInput) { in.Status = "published" }, ""},
		{"bad tag", func(in *postInput) { in.Tags = []string{"Has Space"} }, "tags must be"},
		{"too many tags", func(in *postInput) {
			in.Tags = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
		}, "too many tags"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := valid()
			tc.mutate(&in)
			got := in.validate()
			if tc.wantErr == "" && got != "" {
				t.Fatalf("unexpected error %q", got)
			}
			if tc.wantErr != "" && !strings.Contains(got, tc.wantErr) {
				t.Fatalf("error = %q, want substring %q", got, tc.wantErr)
			}
		})
	}
}

func TestPostInputValidateNormalizesTags(t *testing.T) {
	in := postInput{Slug: "x", Title: "t", Tags: []string{" Go ", "go", "", "Security"}}
	if msg := in.validate(); msg != "" {
		t.Fatal(msg)
	}
	if got := strings.Join(in.Tags, ","); got != "go,security" {
		t.Errorf("tags = %q, want go,security", got)
	}
	if in.Status != "draft" {
		t.Errorf("status = %q, want draft", in.Status)
	}
}

func TestUUIDRe(t *testing.T) {
	for id, want := range map[string]bool{
		"7f9c2ba4-e88f-4a6b-9d3e-1b2c3d4e5f60": true,
		"not-a-uuid":                           false,
		"7f9c2ba4e88f4a6b9d3e1b2c3d4e5f60":     false,
		"' OR 1=1 --":                          false,
		"":                                     false,
	} {
		if uuidRe.MatchString(id) != want {
			t.Errorf("uuidRe(%q) = %v, want %v", id, !want, want)
		}
	}
}
