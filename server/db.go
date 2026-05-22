package server

import (
	"os"
	"path/filepath"
)

type DbPathResult struct {
	Path   string
	Source string
	Exists bool
}

func ResolveWorkspaceDatabase(cwd string) *DbPathResult {
	if v := os.Getenv("BEADS_DIR"); v != "" {
		p := absFrom(v, cwd)
		return &DbPathResult{Path: p, Source: "env", Exists: dirExists(p)}
	}

	if v := os.Getenv("BEADS_DB"); v != "" {
		p := absFrom(v, cwd)
		return &DbPathResult{Path: p, Source: "env", Exists: fileExists(p)}
	}

	beadsDir := findNearestBeadsDir(cwd)
	if beadsDir != "" {
		return &DbPathResult{
			Path:   beadsDir,
			Source: "nearest",
			Exists: true,
		}
	}

	homeDir, _ := os.UserHomeDir()
	homeBeads := filepath.Join(homeDir, ".beads")
	return &DbPathResult{Path: homeBeads, Source: "home-default", Exists: dirExists(homeBeads)}
}

func findNearestBeadsDir(start string) string {
	dir := filepath.Clean(start)
	for i := 0; i < 100; i++ {
		beadsDir := filepath.Join(dir, ".beads")
		if dirExists(beadsDir) {
			metadata := filepath.Join(beadsDir, "metadata.json")
			if fileExists(metadata) {
				return beadsDir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func absFrom(p, cwd string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(cwd, p)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
