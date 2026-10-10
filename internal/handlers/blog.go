package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dillonthompson/dillonthompson.com/internal/markdown"
	"github.com/dillonthompson/dillonthompson.com/internal/repository"
	"github.com/gin-gonic/gin"
)

//go:embed blogassets/templates/*.html
var blogTemplateFS embed.FS

//go:embed blogassets/blog.css blogassets/theme.js blogassets/terminal.js blogassets/geist-latin.woff2
var blogStaticFS embed.FS

// Public pages are cached briefly at the browser and at the Cloudflare edge
// (cache rule in deploy/tofu/cache.tf honors this header). There is
// deliberately no stale-while-revalidate window and no purge step: an
// unpublished or deleted post stays reachable for at most ~2 minutes (60s edge
// + 60s browser), while a new post appears just as quickly.
const (
	htmlCacheControl   = "public, max-age=60, s-maxage=60"
	assetCacheControl  = "public, max-age=3600"
	blogDateLayout     = "January 2, 2006"
	defaultDescription = "Notes on building secure, high-scale systems — Go, TypeScript, and the infrastructure in between."
)

// dbTimeout bounds public DB reads so a Neon cold start can't pin a request
// (and its goroutine) open indefinitely.
const dbTimeout = 8 * time.Second

func dbContext(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), dbTimeout)
}

type BlogHandler struct {
	queries *repository.Queries
	siteURL string
	tmpl    *template.Template
	assets  map[string]blogAsset
}

// blogAsset is a static file served under /blog/assets/.
type blogAsset struct {
	contentType string
	body        []byte
	etag        string // computed once; the bytes never change
}

func newBlogAsset(contentType string, body []byte) blogAsset {
	return blogAsset{contentType: contentType, body: body, etag: etagFor(body)}
}

func etagFor(body []byte) string {
	sum := sha256.Sum256(body)
	return `"` + hex.EncodeToString(sum[:8]) + `"`
}

func NewBlogHandler(queries *repository.Queries, siteURL string) (*BlogHandler, error) {
	tmpl, err := template.ParseFS(blogTemplateFS, "blogassets/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing blog templates: %w", err)
	}
	syntax, err := markdown.SyntaxCSS()
	if err != nil {
		return nil, err
	}
	assets := map[string]blogAsset{
		"syntax.css": newBlogAsset("text/css; charset=utf-8", []byte(syntax)),
	}
	for name, contentType := range map[string]string{
		"blog.css":          "text/css; charset=utf-8",
		"theme.js":          "text/javascript; charset=utf-8",
		"terminal.js":       "text/javascript; charset=utf-8",
		"geist-latin.woff2": "font/woff2",
	} {
		body, err := blogStaticFS.ReadFile("blogassets/" + name)
		if err != nil {
			return nil, fmt.Errorf("reading blog asset %s: %w", name, err)
		}
		assets[name] = newBlogAsset(contentType, body)
	}
	return &BlogHandler{
		queries: queries,
		siteURL: strings.TrimRight(siteURL, "/"),
		tmpl:    tmpl,
		assets:  assets,
	}, nil
}

// pageData is the superset of fields the templates use.
type pageData struct {
	Title        string
	Description  string
	Canonical    string
	SiteURL      string
	OGType       string
	PublishedISO string
	Posts        []postSummary
	Post         *postView
}

type postSummary struct {
	Slug             string
	Title            string
	Description      string
	Tags             []string
	PublishedISO     string
	PublishedDisplay string
}

type postView struct {
	postSummary
	BodyHTML       template.HTML
	ReadingMinutes int
}

func (h *BlogHandler) base(title, description, path, ogType string) pageData {
	if description == "" {
		description = defaultDescription
	}
	return pageData{
		Title:       title,
		Description: description,
		Canonical:   h.siteURL + path,
		SiteURL:     h.siteURL,
		OGType:      ogType,
	}
}

func summary(slug, title, description string, tags []string, published sql.NullTime) postSummary {
	s := postSummary{Slug: slug, Title: title, Description: description, Tags: tags}
	if published.Valid {
		s.PublishedISO = published.Time.UTC().Format(time.RFC3339)
		s.PublishedDisplay = published.Time.UTC().Format(blogDateLayout)
	}
	return s
}

func (h *BlogHandler) Index(c *gin.Context) {
	ctx, cancel := dbContext(c)
	defer cancel()

	rows, err := h.queries.ListPublishedPosts(ctx)
	if err != nil {
		slog.Error("failed to list posts", "error", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	data := h.base("Blog — Dillon Thompson", "", "/blog", "website")
	for _, r := range rows {
		data.Posts = append(data.Posts, summary(r.Slug, r.Title, r.Description, r.Tags, r.PublishedAt))
	}
	h.renderHTML(c, http.StatusOK, "index.html", data)
}

func (h *BlogHandler) Post(c *gin.Context) {
	ctx, cancel := dbContext(c)
	defer cancel()

	slug := c.Param("slug")
	if len(slug) > maxSlugLen || !slugRe.MatchString(slug) {
		// Scanner noise (/blog/aaa.php etc.) shouldn't cost a DB round trip,
		// least of all one that wakes a suspended Neon instance.
		h.notFound(c)
		return
	}
	post, err := h.queries.GetPublishedPostBySlug(ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		h.notFound(c)
		return
	}
	if err != nil {
		slog.Error("failed to fetch post", "slug", slug, "error", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	body, err := markdown.Render(post.BodyMd)
	if err != nil {
		slog.Error("failed to render post", "slug", slug, "error", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	data := h.base(post.Title+" — Dillon Thompson", post.Description, "/blog/"+post.Slug, "article")
	s := summary(post.Slug, post.Title, post.Description, post.Tags, post.PublishedAt)
	data.PublishedISO = s.PublishedISO
	data.Post = &postView{
		postSummary:    s,
		BodyHTML:       body,
		ReadingMinutes: markdown.ReadingMinutes(post.BodyMd),
	}
	h.renderHTML(c, http.StatusOK, "post.html", data)
}

func (h *BlogHandler) notFound(c *gin.Context) {
	h.renderHTML(c, http.StatusNotFound, "notfound.html", h.base("Not found — Dillon Thompson", "", "/blog", "website"))
}

// renderHTML renders to a buffer first so a template error can't leave a
// half-written 200 on the wire, and attaches a content-hash ETag.
func (h *BlogHandler) renderHTML(c *gin.Context, status int, name string, data pageData) {
	var buf bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		slog.Error("failed to render template", "template", name, "error", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}
	if status == http.StatusOK {
		h.serveCached(c, "text/html; charset=utf-8", htmlCacheControl, buf.Bytes())
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Data(status, "text/html; charset=utf-8", buf.Bytes())
}

func (h *BlogHandler) serveCached(c *gin.Context, contentType, cacheControl string, body []byte) {
	h.serveWithETag(c, contentType, cacheControl, body, etagFor(body))
}

func (h *BlogHandler) serveWithETag(c *gin.Context, contentType, cacheControl string, body []byte, etag string) {
	c.Header("ETag", etag)
	c.Header("Cache-Control", cacheControl)
	if etagMatches(c.GetHeader("If-None-Match"), etag) {
		c.Status(http.StatusNotModified)
		return
	}
	c.Data(http.StatusOK, contentType, body)
}

// etagMatches implements the weak comparison If-None-Match requires. Caddy's
// `encode` handler (and Cloudflare) turn strong ETags into weak ones, so
// revalidating clients send W/"..." and sometimes comma-separated lists.
func etagMatches(header, etag string) bool {
	header = strings.TrimSpace(header)
	if header == "" {
		return false
	}
	if header == "*" {
		return true
	}
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimPrefix(strings.TrimSpace(candidate), "W/")
		if candidate == etag {
			return true
		}
	}
	return false
}

// Asset serves one of the embedded static files by its /blog/assets/ name.
// Names are looked up in a fixed map, so no request value ever reaches a
// filesystem path.
func (h *BlogHandler) Asset(name string) gin.HandlerFunc {
	a := h.assets[name]
	return func(c *gin.Context) {
		h.serveWithETag(c, a.contentType, assetCacheControl, a.body, a.etag)
	}
}

// --- RSS ---------------------------------------------------------------------

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string     `xml:"title"`
	Link          string     `xml:"link"`
	Description   string     `xml:"description"`
	Language      string     `xml:"language"`
	LastBuildDate string     `xml:"lastBuildDate,omitempty"`
	AtomLink      rssAtom    `xml:"atom:link"`
	Items         []rssEntry `xml:"item"`
}

type rssAtom struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssEntry struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func (h *BlogHandler) RSS(c *gin.Context) {
	ctx, cancel := dbContext(c)
	defer cancel()

	rows, err := h.queries.ListPublishedPosts(ctx)
	if err != nil {
		slog.Error("failed to list posts for rss", "error", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	feed := rssFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:       "Dillon Thompson",
			Link:        h.siteURL + "/blog",
			Description: defaultDescription,
			Language:    "en-us",
			AtomLink:    rssAtom{Href: h.siteURL + "/rss.xml", Rel: "self", Type: "application/rss+xml"},
		},
	}
	for i, r := range rows {
		if i == 0 && r.PublishedAt.Valid {
			feed.Channel.LastBuildDate = r.PublishedAt.Time.UTC().Format(time.RFC1123Z)
		}
		link := h.siteURL + "/blog/" + r.Slug
		feed.Channel.Items = append(feed.Channel.Items, rssEntry{
			Title:       r.Title,
			Link:        link,
			GUID:        link,
			PubDate:     r.PublishedAt.Time.UTC().Format(time.RFC1123Z),
			Description: r.Description,
		})
	}

	out, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		slog.Error("failed to marshal rss", "error", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}
	h.serveCached(c, "application/rss+xml; charset=utf-8", htmlCacheControl, append([]byte(xml.Header), out...))
}

// --- Sitemap -----------------------------------------------------------------

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

func (h *BlogHandler) Sitemap(c *gin.Context) {
	ctx, cancel := dbContext(c)
	defer cancel()

	rows, err := h.queries.ListPublishedPosts(ctx)
	if err != nil {
		slog.Error("failed to list posts for sitemap", "error", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	set := sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs: []sitemapURL{
			{Loc: h.siteURL + "/"},
			{Loc: h.siteURL + "/about"},
			{Loc: h.siteURL + "/blog"},
		},
	}
	for _, r := range rows {
		set.URLs = append(set.URLs, sitemapURL{
			Loc:     h.siteURL + "/blog/" + r.Slug,
			LastMod: r.UpdatedAt.UTC().Format("2006-01-02"),
		})
	}

	out, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		slog.Error("failed to marshal sitemap", "error", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}
	h.serveCached(c, "application/xml; charset=utf-8", htmlCacheControl, append([]byte(xml.Header), out...))
}
