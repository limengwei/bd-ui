package server

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type DbWatcher struct {
	mu            sync.Mutex
	watcher       *fsnotify.Watcher
	currentPath   string
	currentDir    string
	currentFile   string
	onChange      func()
	debounceMs    time.Duration
	cooldownMs    time.Duration
	cooldownUntil time.Time
	timer         *time.Timer
}

func WatchDb(rootDir string, onChange func(), explicitDb ...string) *DbWatcher {
	debounceMs := 250 * time.Millisecond
	cooldownMs := 1000 * time.Millisecond

	w := &DbWatcher{
		debounceMs: debounceMs,
		cooldownMs: cooldownMs,
		onChange:   onChange,
	}

	var db string
	if len(explicitDb) > 0 {
		db = explicitDb[0]
	}

	w.bind(rootDir, db)
	return w
}

func (w *DbWatcher) bind(rootDir, explicitDb string) {
	opts := &DbPathResult{}
	if explicitDb != "" {
		opts = &DbPathResult{Path: explicitDb, Source: "flag", Exists: fileExists(explicitDb)}
	} else {
		opts = ResolveWorkspaceDatabase(rootDir)
	}

	w.currentPath = opts.Path
	if isDir(opts.Path) {
		w.currentDir = opts.Path
		w.currentFile = "issues.jsonl"
	} else {
		w.currentDir = filepath.Dir(opts.Path)
		w.currentFile = filepath.Base(opts.Path)
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	w.watcher = fsw

	if err := w.watcher.Add(w.currentDir); err != nil {
		return
	}

	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if w.currentFile != "" && event.Name != "" && filepath.Base(event.Name) != w.currentFile {
					continue
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					if time.Now().Before(w.cooldownUntil) {
						continue
					}
					w.schedule()
				}
			case _, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
}

func (w *DbWatcher) schedule() {
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(w.debounceMs, func() {
		w.onChange()
		w.cooldownUntil = time.Now().Add(w.cooldownMs)
	})
}

func (w *DbWatcher) Rebind(rootDir string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	newResolved := ResolveWorkspaceDatabase(rootDir)
	if newResolved.Path != w.currentPath {
		if w.watcher != nil {
			w.watcher.Close()
		}
		w.cooldownUntil = time.Time{}
		w.bind(rootDir, "")
	}
}

func (w *DbWatcher) Close() {
	if w.timer != nil {
		w.timer.Stop()
	}
	if w.watcher != nil {
		w.watcher.Close()
	}
}

func (w *DbWatcher) Path() string {
	return w.currentPath
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
