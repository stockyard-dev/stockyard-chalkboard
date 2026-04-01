package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/stockyard-dev/stockyard-chalkboard/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	port   int
	limits Limits
}

func New(db *store.DB, port int, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), port: port, limits: limits}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /api/pages", s.handleCreatePage)
	s.mux.HandleFunc("GET /api/pages", s.handleListPages)
	s.mux.HandleFunc("GET /api/pages/{id}", s.handleGetPage)
	s.mux.HandleFunc("PUT /api/pages/{id}", s.handleUpdatePage)
	s.mux.HandleFunc("DELETE /api/pages/{id}", s.handleDeletePage)
	s.mux.HandleFunc("GET /api/pages/{id}/versions", s.handleListVersions)
	s.mux.HandleFunc("GET /api/search", s.handleSearch)

	// Public wiki view
	s.mux.HandleFunc("GET /wiki/", s.handleWikiPage)
	s.mux.HandleFunc("GET /wiki/{slug}", s.handleWikiPage)

	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /ui", s.handleUI)
	s.mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"product": "stockyard-chalkboard", "version": "0.1.0"})
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("[chalkboard] listening on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) handleCreatePage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string `json:"title"`
		Slug     string `json:"slug"`
		Content  string `json:"content"`
		ParentID string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		writeJSON(w, 400, map[string]string{"error": "title is required"})
		return
	}
	if s.limits.MaxPages > 0 && LimitReached(s.limits.MaxPages, s.db.TotalPages()) {
		writeJSON(w, 402, map[string]string{"error": fmt.Sprintf("free tier limit: %d pages — upgrade to Pro", s.limits.MaxPages), "upgrade": "https://stockyard.dev/chalkboard/"})
		return
	}
	p, err := s.db.CreatePage(req.Title, req.Slug, req.Content, req.ParentID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"page": p, "url": "/wiki/" + p.Slug})
}

func (s *Server) handleListPages(w http.ResponseWriter, r *http.Request) {
	pages, _ := s.db.ListPages()
	if pages == nil {
		pages = []store.Page{}
	}
	writeJSON(w, 200, map[string]any{"pages": pages, "count": len(pages)})
}

func (s *Server) handleGetPage(w http.ResponseWriter, r *http.Request) {
	p, err := s.db.GetPage(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "page not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"page": p})
}

func (s *Server) handleUpdatePage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetPage(id); err != nil {
		writeJSON(w, 404, map[string]string{"error": "page not found"})
		return
	}
	var req struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	p, _ := s.db.UpdatePage(id, req.Title, req.Content)
	writeJSON(w, 200, map[string]any{"page": p})
}

func (s *Server) handleDeletePage(w http.ResponseWriter, r *http.Request) {
	s.db.DeletePage(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Server) handleListVersions(w http.ResponseWriter, r *http.Request) {
	if !s.limits.VersionHistory {
		writeJSON(w, 402, map[string]string{"error": "version history requires Pro", "upgrade": "https://stockyard.dev/chalkboard/"})
		return
	}
	versions, _ := s.db.ListVersions(r.PathValue("id"))
	if versions == nil {
		versions = []store.PageVersion{}
	}
	writeJSON(w, 200, map[string]any{"versions": versions, "count": len(versions)})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, 400, map[string]string{"error": "q parameter required"})
		return
	}
	pages, _ := s.db.SearchPages(q)
	if pages == nil {
		pages = []store.Page{}
	}
	writeJSON(w, 200, map[string]any{"results": pages, "count": len(pages), "query": q})
}

func (s *Server) handleWikiPage(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		// Index page — list all pages
		pages, _ := s.db.ListPages()
		var list strings.Builder
		for _, p := range pages {
			list.WriteString(fmt.Sprintf(`<div style="margin-bottom:.8rem"><a href="/wiki/%s" style="font-size:1.05rem;color:#e8753a;text-decoration:none">%s</a><div style="font-size:.7rem;color:#7a7060;margin-top:.2rem">Updated %s</div></div>`, he(p.Slug), he(p.Title), p.UpdatedAt[:10]))
		}
		if len(pages) == 0 {
			list.WriteString(`<p style="color:#7a7060;text-align:center;padding:2rem;font-style:italic">No pages yet.</p>`)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, wikiTemplate, "Wiki", "Pages", list.String())
		return
	}
	p, err := s.db.GetPageBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Simple rendering
	content := he(p.Content)
	content = strings.ReplaceAll(content, "\n\n", "</p><p style=\"margin-bottom:1rem;line-height:1.7;color:#bfb5a3\">")
	content = "<p style=\"margin-bottom:1rem;line-height:1.7;color:#bfb5a3\">" + content + "</p>"

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, wikiTemplate, he(p.Title), he(p.Title), content)
}

func he(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.db.Stats()) }
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

const wikiTemplate = `<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s — Chalkboard</title>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:wght@400;700&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
<style>body{background:#1a1410;color:#f0e6d3;font-family:'Libre Baskerville',serif;margin:0;min-height:100vh}
.container{max-width:700px;margin:0 auto;padding:2rem 1.5rem}
h1{font-size:1.4rem;margin-bottom:1.5rem;padding-bottom:.8rem;border-bottom:2px solid #8b3d1a}
a{color:#e8753a;text-decoration:none}a:hover{color:#d4a843}
.back{font-family:'JetBrains Mono',monospace;font-size:.72rem;color:#a0845c;margin-bottom:1.5rem;display:block}
.footer{text-align:center;margin-top:2rem;font-size:.55rem;color:#7a7060;font-family:'JetBrains Mono',monospace}
.footer a{color:#e8753a}
</style></head><body><div class="container">
<a href="/wiki/" class="back">← All pages</a>
<h1>%s</h1>
%s
<div class="footer">Powered by <a href="https://stockyard.dev/chalkboard/">Stockyard Chalkboard</a></div>
</div></body></html>`
