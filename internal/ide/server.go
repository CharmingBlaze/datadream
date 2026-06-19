package ide

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"time"
)

//go:embed web/*
var webFS embed.FS

// WebAssets returns the embedded IDE frontend filesystem (web/ subtree).
func WebAssets() (fs.FS, error) {
	return fs.Sub(webFS, "web")
}

// Server hosts the DataDream web IDE over HTTP.
type Server struct {
	Port int
	svc  *Service
}

// NewServer creates an HTTP IDE server for the given project root.
func NewServer(root string, port int) (*Server, error) {
	svc, err := NewService(root)
	if err != nil {
		return nil, err
	}
	return &Server{Port: port, svc: svc}, nil
}

// Run starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/version", s.handleVersion)
	mux.HandleFunc("GET /api/tree", s.handleTree)
	mux.HandleFunc("GET /api/search", s.handleSearch)
	mux.HandleFunc("GET /api/read", s.handleRead)
	mux.HandleFunc("POST /api/write", s.handleWrite)
	mux.HandleFunc("POST /api/new", s.handleNew)
	mux.HandleFunc("POST /api/check", s.handleCheck)
	mux.HandleFunc("POST /api/build", s.handleBuild)
	mux.HandleFunc("POST /api/run", s.handleRun)
	mux.HandleFunc("GET /api/doctor", s.handleDoctor)

	web, err := WebAssets()
	if err != nil {
		return err
	}
	mux.Handle("/", noCacheFileServer(http.FileServer(http.FS(web))))

	addr := fmt.Sprintf("127.0.0.1:%d", s.Port)
	srv := &http.Server{Addr: addr, Handler: withCORS(mux)}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	fmt.Printf("DataDream Studio running at http://%s\n", addr)
	fmt.Printf("Project root: %s\n", s.svc.Root())
	fmt.Println("Press Ctrl+C to stop.")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func noCacheFileServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.svc.Version())
}

func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	node, err := s.svc.Tree(r.URL.Query().Get("path"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, node)
}

func (s *Server) handleRead(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	file, err := s.svc.Read(path)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, file)
}

type writeRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (s *Server) handleWrite(w http.ResponseWriter, r *http.Request) {
	var req writeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := s.svc.Write(req.Path, req.Content); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

type newFileRequest struct {
	Path     string `json:"path"`
	Template string `json:"template"`
}

func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	var req newFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	file, err := s.svc.NewFile(req.Path, req.Template)
	if err != nil {
		code := http.StatusInternalServerError
		if err.Error() == "file already exists" {
			code = http.StatusConflict
		} else if err.Error() == "path must end with .dd" || err.Error() == "invalid path" {
			code = http.StatusBadRequest
		}
		writeError(w, code, err.Error())
		return
	}
	writeJSON(w, file)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.svc.Search(r.URL.Query().Get("q")))
}

func (s *Server) handleDoctor(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.svc.Doctor())
}

type sourceRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (s *Server) handleCheck(w http.ResponseWriter, r *http.Request) {
	var req sourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	result, err := s.svc.Check(req.Path, req.Content)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, result)
}

func (s *Server) handleBuild(w http.ResponseWriter, r *http.Request) {
	var req struct {
		sourceRequest
		Output  string `json:"output"`
		Release bool   `json:"release"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	writeJSON(w, s.svc.Build(BuildRequest{
		Path: req.Path, Content: req.Content, Output: req.Output, Release: req.Release,
	}))
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	var req sourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	writeJSON(w, s.svc.Run(req.Path, req.Content))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// OpenBrowser tries to open the IDE URL in the default browser.
func OpenBrowser(url string) {
	var cmd *exec.Cmd
	switch {
	case fileExists("/usr/bin/xdg-open"):
		cmd = exec.Command("xdg-open", url)
	case fileExists("/usr/bin/open"):
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("cmd", "/c", "start", url)
	}
	_ = cmd.Start()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
