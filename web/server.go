package web

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"

	"goshare/config"
	"goshare/logger"
	"goshare/storage"
)

//go:embed templates/*
var templateFS embed.FS

type Server struct {
	cfg     *config.Config
	manager *storage.Manager
	mux     *http.ServeMux
}

func NewServer(cfg *config.Config, manager *storage.Manager) *Server {
	s := &Server{
		cfg:     cfg,
		manager: manager,
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Server.Port)
	logger.Info("Starting web server on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/login", s.handleLogin)
	s.mux.HandleFunc("/logout", s.handleLogout)
	s.mux.HandleFunc("/icon.png", s.handleIcon)

	s.mux.HandleFunc("/", s.requireAuth(s.handleIndex))
	s.mux.HandleFunc("/api/upload", s.requireAuthAPI(s.handleUpload))
	s.mux.HandleFunc("/api/download", s.requireAuthAPI(s.handleDownload))
	s.mux.HandleFunc("/api/delete", s.requireAuthAPI(s.handleDelete))
}

var sessionToken string

func init() {
	b := make([]byte, 32)
	rand.Read(b)
	sessionToken = hex.EncodeToString(b)
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.cfg.Auth.Enabled {
			next(w, r)
			return
		}
		cookie, err := r.Cookie("goshare_session")
		if err != nil || cookie.Value != sessionToken {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

func (s *Server) requireAuthAPI(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.cfg.Auth.Enabled {
			next(w, r)
			return
		}
		cookie, err := r.Cookie("goshare_session")
		if err != nil || cookie.Value != sessionToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.Auth.Enabled {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFS(templateFS, "templates/login.html")
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == s.cfg.Auth.Username && password == s.cfg.Auth.Password {
			http.SetCookie(w, &http.Cookie{
				Name:     "goshare_session",
				Value:    sessionToken,
				Path:     "/",
				HttpOnly: true,
			})
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		tmpl, _ := template.ParseFS(templateFS, "templates/login.html")
		tmpl.Execute(w, map[string]string{"Error": "Invalid credentials"})
		return
	}
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "goshare_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *Server) handleIcon(w http.ResponseWriter, r *http.Request) {
	data, err := templateFS.ReadFile("templates/icon.png")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(data)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

type TemplateData struct {
	TotalFiles  int
	TotalSize   int64
	AuthEnabled bool
	Files       []storage.FileMeta
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	funcMap := template.FuncMap{
		"formatBytes": formatBytes,
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "Never"
			}
			return t.Format("2006-01-02 15:04:05")
		},
		"timeUntil": func(t time.Time) string {
			d := time.Until(t)
			if d < 0 {
				return "Expired"
			}
			if d < time.Minute {
				return "Less than a minute"
			}
			if d < time.Hour {
				return fmt.Sprintf("%d mins", int(d.Minutes()))
			}
			if d < 24*time.Hour {
				return fmt.Sprintf("%d hours", int(d.Hours()))
			}
			return fmt.Sprintf("%d days", int(d.Hours()/24))
		},
	}

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFS(templateFS, "templates/index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}

	files := s.manager.GetAllFiles()
	sort.Slice(files, func(i, j int) bool {
		return files[i].UploadedAt.After(files[j].UploadedAt)
	})

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	data := TemplateData{
		TotalFiles:  len(files),
		TotalSize:   totalSize,
		AuthEnabled: s.cfg.Auth.Enabled,
		Files:       files,
	}

	if err := tmpl.Execute(w, data); err != nil {
		logger.Error("Error executing template: %v", err)
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.Storage.MaxUploadSize)
	if err := r.ParseMultipartForm(s.cfg.Storage.MaxUploadSize); err != nil {
		http.Error(w, "File too large or malformed request", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	retention := r.FormValue("retention")
	var duration time.Duration
	switch retention {
	case "1h":
		duration = 1 * time.Hour
	case "12h":
		duration = 12 * time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "168h":
		duration = 168 * time.Hour
	default:
		duration = 24 * time.Hour // default 1 day
	}

	expiresAt := time.Now().Add(duration)

	id, err := s.manager.SaveFile(header.Filename, file, header.Size, expiresAt)
	if err != nil {
		logger.Error("Failed to save file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	logger.Info("File uploaded: %s (ID: %s), Expires at: %s", header.Filename, id, expiresAt.Format(time.RFC3339))
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing file ID", http.StatusBadRequest)
		return
	}

	meta, ok := s.manager.GetFileMeta(id)
	if !ok {
		http.Error(w, "File not found or expired", http.StatusNotFound)
		return
	}

	filePath := s.manager.GetFilePath(id)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", meta.OriginalName))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, filePath)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing file ID", http.StatusBadRequest)
		return
	}

	if err := s.manager.DeleteFile(id); err != nil {
		logger.Error("Failed to delete file %s: %v", id, err)
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	logger.Info("Manually deleted file ID %s", id)
	w.WriteHeader(http.StatusOK)
}
