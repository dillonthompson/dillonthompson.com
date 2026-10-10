package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dillonthompson/dillonthompson.com/internal/markdown"
	"github.com/dillonthompson/dillonthompson.com/internal/middleware"
	"github.com/dillonthompson/dillonthompson.com/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// Limits keep a runaway paste or a bug from filling the table, and bound the
// work the preview endpoint does per request.
const (
	maxBodyBytes     = 200 * 1024
	maxRequestBytes  = 256 * 1024
	maxTitleRunes    = 200
	maxDescRunes     = 300
	maxSlugLen       = 100
	maxTags          = 10
	maxTagLen        = 30
	statusDraft      = "draft"
	statusPublished  = "published"
	pgUniqueViolated = "23505"
	pgCheckViolated  = "23514"
)

var (
	slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	tagRe  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// AdminPostsHandler is the authenticated CRUD API behind the /admin editor.
// It must only be mounted behind middleware.NewAccessAuth and
// middleware.RequireSameOrigin.
type AdminPostsHandler struct {
	queries *repository.Queries
}

func NewAdminPostsHandler(queries *repository.Queries) *AdminPostsHandler {
	return &AdminPostsHandler{queries: queries}
}

// postInput is the client-supplied shape for create and update.
type postInput struct {
	// UpdatedAt is the updated_at of the version the client loaded. Required for
	// updates (optimistic concurrency); ignored on create.
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	BodyMD      string     `json:"body_md"`
	Tags        []string   `json:"tags"`
	Status      string     `json:"status"`
}

// postDTO is the API representation of a post. Body is omitted from list
// responses.
type postDTO struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	BodyMD      *string    `json:"body_md,omitempty"`
	Tags        []string   `json:"tags"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func nullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

func tags(t []string) []string {
	if t == nil {
		return []string{}
	}
	return t
}

func fullDTO(p repository.Post) postDTO {
	body := p.BodyMd
	return postDTO{
		ID: p.ID, Slug: p.Slug, Title: p.Title, Description: p.Description, BodyMD: &body,
		Tags: tags(p.Tags), Status: p.Status, PublishedAt: nullTime(p.PublishedAt),
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

// validate normalizes the input in place and returns a user-facing error
// message, or "" when the input is acceptable.
func (in *postInput) validate() string {
	in.Slug = strings.TrimSpace(in.Slug)
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.Status = strings.TrimSpace(in.Status)
	if in.Status == "" {
		in.Status = statusDraft
	}

	if strings.ContainsRune(in.Title+in.Slug+in.Description+in.BodyMD+strings.Join(in.Tags, ""), 0) {
		// Postgres TEXT can't store NUL bytes.
		return "text contains an invalid character"
	}

	switch {
	case in.Title == "":
		return "title is required"
	case utf8.RuneCountInString(in.Title) > maxTitleRunes:
		return "title is too long"
	case in.Slug == "" || len(in.Slug) > maxSlugLen || !slugRe.MatchString(in.Slug):
		return "slug must be lowercase letters, numbers and single hyphens"
	case utf8.RuneCountInString(in.Description) > maxDescRunes:
		return "description is too long"
	case len(in.BodyMD) > maxBodyBytes:
		return "body is too long"
	case in.Status != statusDraft && in.Status != statusPublished:
		return "status must be draft or published"
	case in.Status == statusPublished && strings.TrimSpace(in.BodyMD) == "":
		return "a published post needs a body"
	}

	seen := make(map[string]struct{}, len(in.Tags))
	clean := make([]string, 0, len(in.Tags))
	for _, t := range in.Tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if len(t) > maxTagLen || !tagRe.MatchString(t) {
			return "tags must be lowercase letters, numbers and hyphens (max 30 characters)"
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		clean = append(clean, t)
	}
	if len(clean) > maxTags {
		return "too many tags (max 10)"
	}
	in.Tags = clean
	return ""
}

func noStore(c *gin.Context) { c.Header("Cache-Control", "no-store") }

// readInput decodes and validates a request body, writing the error response
// itself on failure.
func readInput(c *gin.Context) (postInput, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	var in postInput
	if err := c.ShouldBindJSON(&in); err != nil {
		writeBindError(c, err)
		return in, false
	}
	if msg := in.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return in, false
	}
	return in, true
}

// writeBindError distinguishes an oversized body (413) from malformed JSON (400).
func writeBindError(c *gin.Context, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request is too large"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
}

// idParam returns the :id route param if it is a UUID. Otherwise it writes a
// 404 (so malformed ids never reach the database) and reports false.
func idParam(c *gin.Context) (string, bool) {
	id := c.Param("id")
	if !uuidRe.MatchString(id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return "", false
	}
	return id, true
}

// writeDBError maps database errors to API responses without leaking details.
func writeDBError(c *gin.Context, op string, err error) {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case pgUniqueViolated:
			c.JSON(http.StatusConflict, gin.H{"error": "slug is already in use"})
			return
		case pgCheckViolated:
			c.JSON(http.StatusBadRequest, gin.H{"error": "post violates a data constraint"})
			return
		}
	}
	slog.Error("admin posts: "+op+" failed", "error", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func (h *AdminPostsHandler) Me(c *gin.Context) {
	noStore(c)
	c.JSON(http.StatusOK, gin.H{"email": middleware.AdminEmail(c)})
}

func (h *AdminPostsHandler) List(c *gin.Context) {
	noStore(c)
	rows, err := h.queries.ListAllPosts(c.Request.Context())
	if err != nil {
		writeDBError(c, "list", err)
		return
	}
	out := make([]postDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, postDTO{
			ID: r.ID, Slug: r.Slug, Title: r.Title, Description: r.Description,
			Tags: tags(r.Tags), Status: r.Status, PublishedAt: nullTime(r.PublishedAt),
			CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *AdminPostsHandler) Get(c *gin.Context) {
	noStore(c)
	id, ok := idParam(c)
	if !ok {
		return
	}
	post, err := h.queries.GetPostByID(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		writeDBError(c, "get", err)
		return
	}
	c.JSON(http.StatusOK, fullDTO(post))
}

func (h *AdminPostsHandler) Create(c *gin.Context) {
	noStore(c)
	in, ok := readInput(c)
	if !ok {
		return
	}
	post, err := h.queries.CreatePost(c.Request.Context(), repository.CreatePostParams{
		Slug: in.Slug, Title: in.Title, Description: in.Description, BodyMd: in.BodyMD, Tags: in.Tags, Status: in.Status,
	})
	if err != nil {
		writeDBError(c, "create", err)
		return
	}
	slog.Info("admin: post created", "admin", middleware.AdminEmail(c), "post_id", post.ID, "slug", post.Slug, "status", post.Status)
	c.JSON(http.StatusCreated, fullDTO(post))
}

func (h *AdminPostsHandler) Update(c *gin.Context) {
	noStore(c)
	id, ok := idParam(c)
	if !ok {
		return
	}
	in, ok := readInput(c)
	if !ok {
		return
	}
	if in.UpdatedAt == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "updated_at is required"})
		return
	}
	post, err := h.queries.UpdatePost(c.Request.Context(), repository.UpdatePostParams{
		ID: id, Slug: in.Slug, Title: in.Title, Description: in.Description, BodyMd: in.BodyMD, Tags: in.Tags, Status: in.Status,
		ExpectedUpdatedAt: *in.UpdatedAt,
	})
	if errors.Is(err, sql.ErrNoRows) {
		// Either the post is gone, or it changed since this client loaded it
		// (another tab or device). Don't silently overwrite the newer version.
		if _, getErr := h.queries.GetPostByID(c.Request.Context(), id); getErr == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "this post was changed since you loaded it (another tab?). Reload to get the latest version."})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		writeDBError(c, "update", err)
		return
	}
	slog.Info("admin: post updated", "admin", middleware.AdminEmail(c), "post_id", post.ID, "slug", post.Slug, "status", post.Status)
	c.JSON(http.StatusOK, fullDTO(post))
}

func (h *AdminPostsHandler) Delete(c *gin.Context) {
	noStore(c)
	id, ok := idParam(c)
	if !ok {
		return
	}
	n, err := h.queries.DeletePost(c.Request.Context(), id)
	if err != nil {
		writeDBError(c, "delete", err)
		return
	}
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	slog.Info("admin: post deleted", "admin", middleware.AdminEmail(c), "post_id", id)
	c.Status(http.StatusNoContent)
}

// Preview renders Markdown with the exact pipeline the public site uses
// (including sanitization), so what the editor shows is what readers get and
// the client never has to be trusted to render untrusted HTML itself.
func (h *AdminPostsHandler) Preview(c *gin.Context) {
	noStore(c)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	var in struct {
		BodyMD string `json:"body_md"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeBindError(c, err)
		return
	}
	if len(in.BodyMD) > maxBodyBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "body is too large"})
		return
	}
	html, err := markdown.Render(in.BodyMD)
	if err != nil {
		slog.Error("admin preview: render failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"html": string(html), "reading_minutes": markdown.ReadingMinutes(in.BodyMD)})
}
