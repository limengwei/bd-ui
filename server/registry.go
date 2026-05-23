package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type InMemoryWorkspace struct {
	Path     string `json:"path"`
	Database string `json:"database"`
	Pid      int    `json:"pid"`
	Version  string `json:"version"`
}

type WorkspaceInfo struct {
	Path     string `json:"path"`
	Database string `json:"database"`
	Pid      int    `json:"pid"`
	Version  string `json:"version"`
}

type PersistedConfig struct {
	Workspaces []WorkspaceInfo `json:"workspaces"`
	BdBinPath  string          `json:"bd_bin_path"`
}

var (
	registryMu sync.Mutex
)

func getConfigFile() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "config.json")
	}
	return "config.json"
}

func loadConfig() *PersistedConfig {
	path := getConfigFile()
	data, err := os.ReadFile(path)
	if err != nil {
		return &PersistedConfig{
			Workspaces: []WorkspaceInfo{},
			BdBinPath:  "",
		}
	}
	var cfg PersistedConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &PersistedConfig{
			Workspaces: []WorkspaceInfo{},
			BdBinPath:  "",
		}
	}
	return &cfg
}

func saveConfig(cfg *PersistedConfig) {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(getConfigFile(), data, 0644)
}

func loadPersistedWorkspaces() map[string]*WorkspaceInfo {
	cfg := loadConfig()
	m := make(map[string]*WorkspaceInfo, len(cfg.Workspaces))
	for _, w := range cfg.Workspaces {
		w := w
		absPath, _ := filepath.Abs(w.Path)
		w.Path = absPath
		m[absPath] = &w
	}
	return m
}

func savePersistedWorkspaces(m map[string]*WorkspaceInfo) {
	cfg := loadConfig()
	var list []WorkspaceInfo
	for _, w := range m {
		list = append(list, *w)
	}
	cfg.Workspaces = list
	saveConfig(cfg)
}

func AddWorkspace(path string) *WorkspaceInfo {
	registryMu.Lock()
	defer registryMu.Unlock()

	absPath, _ := filepath.Abs(path)
	db := ResolveWorkspaceDatabase(absPath)
	info := &WorkspaceInfo{
		Path:     absPath,
		Database: db.Path,
		Version:  "manual",
	}

	m := loadPersistedWorkspaces()
	m[absPath] = info
	savePersistedWorkspaces(m)
	return info
}

func RemoveWorkspace(path string) {
	registryMu.Lock()
	defer registryMu.Unlock()

	absPath, _ := filepath.Abs(path)
	m := loadPersistedWorkspaces()
	delete(m, absPath)
	savePersistedWorkspaces(m)
}

func GetAvailableWorkspaces() []WorkspaceInfo {
	registryMu.Lock()
	defer registryMu.Unlock()

	m := loadPersistedWorkspaces()
	var result []WorkspaceInfo
	for _, w := range m {
		result = append(result, *w)
	}
	if len(result) == 0 {
		return []WorkspaceInfo{}
	}
	return result
}
