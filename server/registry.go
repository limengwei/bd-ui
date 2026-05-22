package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type RegistryEntry struct {
	WorkspacePath string `json:"workspace_path"`
	SocketPath    string `json:"socket_path"`
	DatabasePath  string `json:"database_path"`
	Pid           int    `json:"pid"`
	Version       string `json:"version"`
	StartedAt     string `json:"started_at"`
}

type InMemoryWorkspace struct {
	Path     string `json:"path"`
	Database string `json:"database"`
	Pid      int    `json:"pid"`
	Version  string `json:"version"`
}

var (
	inMemoryMu        sync.Mutex
	inMemoryWorkspaces = make(map[string]*InMemoryWorkspace)
)

func RegisterWorkspace(path, database string) {
	inMemoryMu.Lock()
	defer inMemoryMu.Unlock()
	absPath, _ := filepath.Abs(path)
	inMemoryWorkspaces[absPath] = &InMemoryWorkspace{
		Path:     absPath,
		Database: database,
		Pid:      os.Getpid(),
		Version:  "dynamic",
	}
}

func GetRegistryPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".beads", "registry.json")
}

func ReadRegistry() []RegistryEntry {
	path := GetRegistryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries []RegistryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	return entries
}

type WorkspaceInfo struct {
	Path     string `json:"path"`
	Database string `json:"database"`
	Pid      int    `json:"pid"`
	Version  string `json:"version"`
}

func GetAvailableWorkspaces() []WorkspaceInfo {
	entries := ReadRegistry()
	var workspaces []WorkspaceInfo

	seen := make(map[string]bool)
	for _, e := range entries {
		absPath, _ := filepath.Abs(e.WorkspacePath)
		workspaces = append(workspaces, WorkspaceInfo{
			Path:     absPath,
			Database: e.DatabasePath,
			Pid:      e.Pid,
			Version:  e.Version,
		})
		seen[absPath] = true
	}

	inMemoryMu.Lock()
	defer inMemoryMu.Unlock()
	for _, ws := range inMemoryWorkspaces {
		if !seen[ws.Path] {
			workspaces = append(workspaces, WorkspaceInfo{
				Path:     ws.Path,
				Database: ws.Database,
				Pid:      ws.Pid,
				Version:  ws.Version,
			})
		}
	}

	if len(workspaces) == 0 {
		return []WorkspaceInfo{}
	}
	return workspaces
}
