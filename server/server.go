package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

type Server struct {
	config   *Config
	wsServer *WsServer
	watcher  *DbWatcher
	staticFS fs.FS
}

func NewServer(config *Config, webFS fs.FS) *Server {
	return &Server{
		config:   config,
		staticFS: webFS,
	}
}

func (sv *Server) Start() error {
	dbResult := ResolveWorkspaceDatabase(sv.config.RootDir)
	if dbResult.Source != "home-default" && dbResult.Exists {
		RegisterWorkspace(sv.config.RootDir, dbResult.Path)
	}

	sv.watcher = WatchDb(sv.config.RootDir, func() {
		log.Println("数据库变更检测 → 触发刷新")
		if sv.wsServer != nil {
			sv.wsServer.ScheduleListRefresh()
		}
	})

	sv.wsServer = NewWsServer(sv.watcher, sv.config.RootDir)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	mux.HandleFunc("/api/register-workspace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Path     string `json:"path"`
			Database string `json:"database"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if body.Path == "" || body.Database == "" {
			http.Error(w, "Missing path or database", http.StatusBadRequest)
			return
		}
		RegisterWorkspace(body.Path, body.Database)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "registered": body.Path})
	})

	mux.HandleFunc("/api/workspaces", func(w http.ResponseWriter, r *http.Request) {
		workspaces := GetAvailableWorkspaces()
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "workspaces": workspaces})
	})

	mux.HandleFunc("/ws", sv.wsServer.HandleWS)

	if sv.staticFS != nil {
		fileServer := http.FileServer(http.FS(sv.staticFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				r.URL.Path = "/index.html"
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	addr := fmt.Sprintf("%s:%d", sv.config.Host, sv.config.Port)
	log.Printf("Beads UI 服务器启动在 %s", sv.config.URL)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	openBrowser(sv.config.URL)
	<-sigChan
	log.Println("正在关闭服务器...")
	sv.watcher.Close()
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}
