// Package markdown renders post bodies to sanitized HTML.
//
// Rendering is two layers on purpose: goldmark turns Markdown into HTML (raw
// HTML in the source is dropped, not passed through), then bluemonday
// allow-lists the result. The author is trusted, but the second pass means a
// compromised admin session or a parser quirk can't turn into stored XSS on
// the public site.
package markdown

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

const (
	lightStyle = "github"
	darkStyle  = "github-dark"
)

var (
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			highlighting.NewHighlighting(
				// Emit CSS classes rather than inline styles: the CSP allows
				// inline styles today, but classes keep light/dark themeable
				// and let bluemonday strip every style attribute.
				highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps()),
	)

	policy = newPolicy()

	wordRe = regexp.MustCompile(`\S+`)
)

func newPolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	// Chroma classes on code blocks; heading ids for anchor links.
	classes := regexp.MustCompile(`^[a-zA-Z0-9 _-]+$`)
	p.AllowAttrs("class").Matching(classes).OnElements("pre", "code", "span", "div", "sup", "li", "a", "ol", "section")
	p.AllowAttrs("id").Matching(regexp.MustCompile(`^[a-zA-Z0-9_:.-]+$`)).OnElements("h1", "h2", "h3", "h4", "h5", "h6", "li", "sup")
	// GFM task lists: <input type="checkbox" checked disabled>. Only static,
	// disabled checkboxes — no other input types.
	p.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
	p.AllowAttrs("checked", "disabled").Matching(regexp.MustCompile(`^(checked|disabled)?$`)).OnElements("input")
	// GFM table alignment is emitted as style="text-align:...". Allow that one
	// property with a fixed set of values; every other style stays stripped.
	p.AllowStyles("text-align").MatchingEnum("left", "right", "center").OnElements("th", "td")
	// ARIA roles goldmark's footnote extension adds.
	p.AllowAttrs("role").Matching(regexp.MustCompile(`^doc-(noteref|backlink|endnotes)$`)).OnElements("a", "div")
	// The author's own outbound links shouldn't be marked nofollow.
	p.RequireNoFollowOnLinks(false)
	p.RequireNoReferrerOnLinks(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)
	return p
}

// Render converts Markdown to sanitized HTML.
func Render(src string) (template.HTML, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return "", fmt.Errorf("rendering markdown: %w", err)
	}
	// The input to Sanitize is untrusted; its output is safe by construction.
	return template.HTML(policy.SanitizeBytes(buf.Bytes())), nil //nolint:gosec
}

// ReadingMinutes estimates reading time at ~200 wpm, minimum 1.
func ReadingMinutes(src string) int {
	words := len(wordRe.FindAllString(src, -1))
	if m := (words + 199) / 200; m > 1 {
		return m
	}
	return 1
}

// SyntaxCSS returns the stylesheet for highlighted code. The two themes are
// scoped so exactly one ever applies, matching how the SPA toggles themes (a
// `dark` class on <html>): light under `:root:not(.dark)`, dark under `.dark`.
//
// Scoping both matters: the light theme defines token colors (e.g. punctuation)
// that the dark theme leaves unset, so an unscoped light rule would keep
// winning in dark mode and render dark text on a dark background.
func SyntaxCSS() (string, error) {
	var out strings.Builder
	if err := writeCSS(&out, lightStyle, ":root:not(.dark) "); err != nil {
		return "", err
	}
	if err := writeCSS(&out, darkStyle, ".dark "); err != nil {
		return "", err
	}
	return out.String(), nil
}

func writeCSS(out *strings.Builder, name, scope string) error {
	style := styles.Get(name)
	if style == nil || style == styles.Fallback {
		return fmt.Errorf("chroma style %q not found", name)
	}
	var buf bytes.Buffer
	if err := chromahtml.New(chromahtml.WithClasses(true)).WriteCSS(&buf, style); err != nil {
		return fmt.Errorf("writing chroma css: %w", err)
	}
	for _, line := range strings.Split(buf.String(), "\n") {
		// Every rule we emit targets `.chroma`. Chroma also emits a bare `.bg`
		// rule for its Background class, which our markup never uses and which
		// couldn't be scoped by the prefix below — drop it.
		if !strings.Contains(line, ".chroma") {
			continue
		}
		out.WriteString(strings.ReplaceAll(line, ".chroma", scope+".chroma"))
		out.WriteString("\n")
	}
	return nil
}
