package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ conn *sql.DB }

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	conn, err := sql.Open("sqlite", filepath.Join(dataDir, "chalkboard.db"))
	if err != nil {
		return nil, err
	}
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA busy_timeout=5000")
	conn.SetMaxOpenConns(4)
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error { return db.conn.Close() }

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS pages (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    content TEXT DEFAULT '',
    parent_id TEXT DEFAULT '',
    sort_order INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_pages_slug ON pages(slug);
CREATE INDEX IF NOT EXISTS idx_pages_parent ON pages(parent_id);

CREATE TABLE IF NOT EXISTS page_versions (
    id TEXT PRIMARY KEY,
    page_id TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT DEFAULT '',
    version INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_ver_page ON page_versions(page_id);
`)
	return err
}

type Page struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	ParentID  string `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (db *DB) CreatePage(title, slug, content, parentID string) (*Page, error) {
	id := "pg_" + genID(8)
	now := time.Now().UTC().Format(time.RFC3339)
	if slug == "" {
		slug = slugify(title)
	}
	_, err := db.conn.Exec("INSERT INTO pages (id,slug,title,content,parent_id,created_at,updated_at) VALUES (?,?,?,?,?,?,?)",
		id, slug, title, content, parentID, now, now)
	if err != nil {
		return nil, err
	}
	db.conn.Exec("INSERT INTO page_versions (id,page_id,title,content,version) VALUES (?,?,?,?,1)", "v_"+genID(6), id, title, content)
	return &Page{ID: id, Slug: slug, Title: title, Content: content, ParentID: parentID, CreatedAt: now, UpdatedAt: now}, nil
}

func (db *DB) ListPages() ([]Page, error) {
	rows, err := db.conn.Query("SELECT id,slug,title,content,parent_id,sort_order,created_at,updated_at FROM pages ORDER BY sort_order, title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Page
	for rows.Next() {
		var p Page
		rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Content, &p.ParentID, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (db *DB) GetPage(id string) (*Page, error) {
	var p Page
	err := db.conn.QueryRow("SELECT id,slug,title,content,parent_id,sort_order,created_at,updated_at FROM pages WHERE id=?", id).
		Scan(&p.ID, &p.Slug, &p.Title, &p.Content, &p.ParentID, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (db *DB) GetPageBySlug(slug string) (*Page, error) {
	var p Page
	err := db.conn.QueryRow("SELECT id,slug,title,content,parent_id,sort_order,created_at,updated_at FROM pages WHERE slug=?", slug).
		Scan(&p.ID, &p.Slug, &p.Title, &p.Content, &p.ParentID, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (db *DB) UpdatePage(id string, title, content *string) (*Page, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if title != nil {
		db.conn.Exec("UPDATE pages SET title=?, updated_at=? WHERE id=?", *title, now, id)
	}
	if content != nil {
		db.conn.Exec("UPDATE pages SET content=?, updated_at=? WHERE id=?", *content, now, id)
	}
	// Save version
	p, err := db.GetPage(id)
	if err != nil {
		return nil, err
	}
	var maxVer int
	db.conn.QueryRow("SELECT COALESCE(MAX(version),0) FROM page_versions WHERE page_id=?", id).Scan(&maxVer)
	db.conn.Exec("INSERT INTO page_versions (id,page_id,title,content,version) VALUES (?,?,?,?,?)",
		"v_"+genID(6), id, p.Title, p.Content, maxVer+1)
	return p, nil
}

func (db *DB) DeletePage(id string) error {
	db.conn.Exec("DELETE FROM page_versions WHERE page_id=?", id)
	_, err := db.conn.Exec("DELETE FROM pages WHERE id=?", id)
	return err
}

func (db *DB) SearchPages(query string) ([]Page, error) {
	q := "%" + query + "%"
	rows, err := db.conn.Query("SELECT id,slug,title,content,parent_id,sort_order,created_at,updated_at FROM pages WHERE title LIKE ? OR content LIKE ? ORDER BY updated_at DESC LIMIT 20", q, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Page
	for rows.Next() {
		var p Page
		rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Content, &p.ParentID, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
		out = append(out, p)
	}
	return out, rows.Err()
}

type PageVersion struct {
	ID        string `json:"id"`
	PageID    string `json:"page_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Version   int    `json:"version"`
	CreatedAt string `json:"created_at"`
}

func (db *DB) ListVersions(pageID string) ([]PageVersion, error) {
	rows, err := db.conn.Query("SELECT id,page_id,title,content,version,created_at FROM page_versions WHERE page_id=? ORDER BY version DESC", pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PageVersion
	for rows.Next() {
		var v PageVersion
		rows.Scan(&v.ID, &v.PageID, &v.Title, &v.Content, &v.Version, &v.CreatedAt)
		out = append(out, v)
	}
	return out, rows.Err()
}

func (db *DB) TotalPages() int {
	var c int
	db.conn.QueryRow("SELECT COUNT(*) FROM pages").Scan(&c)
	return c
}

func (db *DB) Stats() map[string]any {
	var pages, versions int
	db.conn.QueryRow("SELECT COUNT(*) FROM pages").Scan(&pages)
	db.conn.QueryRow("SELECT COUNT(*) FROM page_versions").Scan(&versions)
	return map[string]any{"pages": pages, "versions": versions}
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			b.WriteRune(c)
		} else if c == ' ' || c == '-' {
			b.WriteByte('-')
		}
	}
	r := b.String()
	for strings.Contains(r, "--") {
		r = strings.ReplaceAll(r, "--", "-")
	}
	return strings.Trim(r, "-")
}

func genID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
