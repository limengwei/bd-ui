package server

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DbPathResult struct {
	Path   string
	Source string
	Exists bool
}

func ResolveDbPath(cwd string) *DbPathResult {
	homeDir, _ := os.UserHomeDir()
	homeDefault := filepath.Join(homeDir, ".beads", "default.db")

	if v := os.Getenv("BEADS_DB"); v != "" {
		p := absFrom(v, cwd)
		return &DbPathResult{Path: p, Source: "env", Exists: fileExists(p)}
	}

	nearest := findNearestBeadsDb(cwd)
	if nearest != "" && normalizePath(nearest) != normalizePath(homeDefault) {
		return &DbPathResult{Path: nearest, Source: "nearest", Exists: fileExists(nearest)}
	}

	return &DbPathResult{Path: homeDefault, Source: "home-default", Exists: fileExists(homeDefault)}
}

func ResolveWorkspaceDatabase(cwd string) *DbPathResult {
	sqliteDb := ResolveDbPath(cwd)
	if sqliteDb.Source != "home-default" {
		return sqliteDb
	}

	metadataPath := findNearestBeadsMetadata(cwd)
	if metadataPath != "" {
		return &DbPathResult{
			Path:   filepath.Dir(metadataPath),
			Source: "metadata",
			Exists: true,
		}
	}

	return sqliteDb
}

func findNearestBeadsMetadata(start string) string {
	dir := filepath.Clean(start)
	for i := 0; i < 100; i++ {
		p := filepath.Join(dir, ".beads", "metadata.json")
		if fileExists(p) {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func findNearestBeadsDb(start string) string {
	dir := filepath.Clean(start)
	for i := 0; i < 100; i++ {
		beadsDir := filepath.Join(dir, ".beads")
		entries, err := os.ReadDir(beadsDir)
		if err == nil {
			var dbs []string
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".db") {
					dbs = append(dbs, e.Name())
				}
			}
			if len(dbs) > 0 {
				sort.Strings(dbs)
				return filepath.Join(beadsDir, dbs[0])
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

func normalizePath(p string) string {
	return filepath.Clean(strings.ToLower(p))
}
