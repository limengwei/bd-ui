package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var bdSerialMu sync.Mutex

var (
	bdBinMu       sync.RWMutex
	bdBinOverride string
)

func init() {
	if cfgPath := LoadBdBinPath(); cfgPath != "" {
		bdBinOverride = cfgPath
	}
}

func GetBdBin() string {
	bdBinMu.RLock()
	if bdBinOverride != "" {
		v := bdBinOverride
		bdBinMu.RUnlock()
		return v
	}
	bdBinMu.RUnlock()
	if v := os.Getenv("BD_BIN"); v != "" {
		return v
	}
	if exe, err := os.Executable(); err == nil {
		name := "bd"
		if runtime.GOOS == "windows" {
			name = "bd.exe"
		}
		candidate := filepath.Join(filepath.Dir(exe), name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "bd"
}

func SetBdBin(path string) {
	bdBinMu.Lock()
	defer bdBinMu.Unlock()
	bdBinOverride = path
	// 保存到配置文件
	SaveBdBinPath(path)
}

type BdResult struct {
	Code   int
	Stdout string
	Stderr string
}

func RunBd(args []string, cwd string) (*BdResult, error) {
	bdSerialMu.Lock()
	defer bdSerialMu.Unlock()

	return runBdUnlocked(args, cwd)
}

func runBdUnlocked(args []string, cwd string) (*BdResult, error) {
	bin := GetBdBin()

	dbPath := ResolveWorkspaceDatabase(cwd)
	env := os.Environ()
	if dbPath.Exists && dbPath.Source != "home-default" {
		env = append(env, fmt.Sprintf("BEADS_DIR=%s", dbPath.Path))
	}

	finalArgs := buildBdArgs(args)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, finalArgs...)
	cmd.Dir = cwd
	cmd.Env = env

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			code = 127
		}
	}

	return &BdResult{
		Code:   code,
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}, nil
}

func buildBdArgs(args []string) []string {
	rawSandbox := strings.ToLower(os.Getenv("BDUI_BD_SANDBOX"))
	sandboxDisabled := rawSandbox == "0" || rawSandbox == "false"

	hasSandbox := false
	for _, a := range args {
		if a == "--sandbox" {
			hasSandbox = true
			break
		}
	}

	if sandboxDisabled || hasSandbox {
		result := make([]string, len(args))
		copy(result, args)
		return result
	}

	return append([]string{"--sandbox"}, args...)
}

func RunBdJson(args []string, cwd string) (int, interface{}, string) {
	result, err := RunBd(args, cwd)
	if err != nil {
		return 127, nil, err.Error()
	}
	if result.Code != 0 {
		return result.Code, nil, result.Stderr
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(result.Stdout), &parsed); err != nil {
		return 0, nil, "Invalid JSON from bd"
	}
	return 0, parsed, ""
}
