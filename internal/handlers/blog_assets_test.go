package handlers

import (
	"bytes"
	"strings"
	"testing"
)

// Asset URLs in rendered pages must carry the file's content hash, so a deploy
// that changes an asset can't be masked by a stale Cloudflare/browser copy.
func TestBlogPagesReferenceHashedAssets(t *testing.T) {
	h, err := NewBlogHandler(nil, "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&buf, "head", pageData{Title: "t"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, name := range []string{"blog.css", "theme.js", "terminal.js", "syntax.css"} {
		want := `/blog/assets/` + name + `?v=` + strings.Trim(h.assets[name].etag, `"`)
		if !strings.Contains(out, want) {
			t.Errorf("head does not reference %s", want)
		}
	}
}
