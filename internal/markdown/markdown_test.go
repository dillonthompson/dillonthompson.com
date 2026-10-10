package markdown

import (
	"strings"
	"testing"
)

func TestRenderStripsDangerousContent(t *testing.T) {
	cases := map[string]string{
		"raw script":      "hello <script>alert(1)</script>",
		"inline handler":  `<img src=x onerror="alert(1)">`,
		"javascript link": "[x](javascript:alert(1))",
		"iframe":          `<iframe src="https://evil.example"></iframe>`,
		"style attr":      `<p style="position:fixed">hi</p>`,
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := Render(src)
			if err != nil {
				t.Fatal(err)
			}
			out := strings.ToLower(string(got))
			for _, bad := range []string{"<script", "onerror", "javascript:", "<iframe", "style="} {
				if strings.Contains(out, bad) {
					t.Errorf("output contains %q: %s", bad, got)
				}
			}
		})
	}
}

func TestRenderKeepsFormatting(t *testing.T) {
	got, err := Render("## Heading\n\n```go\nfmt.Println(\"hi\")\n```\n\n[ext](https://example.com)\n")
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	for _, want := range []string{`<h2 id="heading">`, `class="chroma"`, `href="https://example.com"`, `noreferrer`} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %s", want, out)
		}
	}
	if strings.Contains(out, "nofollow") {
		t.Errorf("author links should not be nofollow: %s", out)
	}
}

func TestRenderGFMFeatures(t *testing.T) {
	src := "- [x] done\n- [ ] todo\n\n| a | b |\n|:-:|--:|\n| 1 | 2 |\n\nref[^1]\n\n[^1]: note\n"
	got, err := Render(src)
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	for _, want := range []string{
		`type="checkbox"`, `checked=""`, `disabled=""`,
		`style="text-align: center"`, `style="text-align: right"`,
		`href="#fn:1"`, `id="fn:1"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %s", want, out)
		}
	}
}

func TestRenderRejectsUnsafeInputsAndStyles(t *testing.T) {
	cases := map[string]string{
		"text input":      `<input type="text" name="x">`,
		"other style":     `<td style="position:fixed;text-align:center">x</td>`,
		"bad align value": `<td style="text-align:expression(alert(1))">x</td>`,
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := Render(src)
			out := strings.ToLower(string(got))
			for _, bad := range []string{`type="text"`, "position", "expression"} {
				if strings.Contains(out, bad) {
					t.Errorf("output contains %q: %s", bad, got)
				}
			}
		})
	}
}

func TestReadingMinutes(t *testing.T) {
	if got := ReadingMinutes("short"); got != 1 {
		t.Errorf("short = %d, want 1", got)
	}
	if got := ReadingMinutes(strings.Repeat("word ", 450)); got != 3 {
		t.Errorf("450 words = %d, want 3", got)
	}
}

func TestSyntaxCSSIsFullyScoped(t *testing.T) {
	css, err := SyntaxCSS()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(css, ":root:not(.dark) .chroma") || !strings.Contains(css, ".dark .chroma") {
		t.Fatal("expected both light and dark scoped rules")
	}
	// Regression: an unscoped rule from one theme leaks into the other (light
	// punctuation color was winning in dark mode). Every selector must carry
	// one of the two scopes.
	for _, line := range strings.Split(css, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		sel := line
		if i := strings.Index(line, "*/"); i >= 0 {
			sel = strings.TrimSpace(line[i+2:])
		}
		if !strings.HasPrefix(sel, ":root:not(.dark) .chroma") && !strings.HasPrefix(sel, ".dark .chroma") {
			t.Errorf("unscoped rule leaks across themes: %q", line)
		}
	}
}
