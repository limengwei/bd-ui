package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type RequestEnvelope struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type ReplyEnvelope struct {
	ID      string      `json:"id"`
	OK      bool        `json:"ok"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	Error   *ReplyError `json:"error,omitempty"`
}

type ReplyError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ConnState struct {
	mu       sync.Mutex
	ShowID   string
	ListSubs map[string]SubSpec
	ListRev  map[string]int
}

type SubSpec struct {
	Key  string
	Spec map[string]interface{}
}

type WsServer struct {
	registry     *SubRegistry
	watcher      *DbWatcher
	upgrader     websocket.Upgrader
	connections  map[*websocket.Conn]*ConnState
	mu           sync.Mutex
	workspace    *WorkspaceState
	refreshMu    sync.Mutex
	refreshTimer *time.Timer
}

type WorkspaceState struct {
	RootDir string
	DbPath  string
}

func NewWsServer(watcher *DbWatcher, rootDir string) *WsServer {
	dbResult := ResolveWorkspaceDatabase(rootDir)
	return &WsServer{
		registry:    NewSubRegistry(),
		watcher:     watcher,
		upgrader:    websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		connections: make(map[*websocket.Conn]*ConnState),
		workspace: &WorkspaceState{
			RootDir: rootDir,
			DbPath:  dbResult.Path,
		},
	}
}

func (s *WsServer) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	state := &ConnState{
		ListSubs: make(map[string]SubSpec),
		ListRev:  make(map[string]int),
	}

	s.mu.Lock()
	s.connections[conn] = state
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.connections, conn)
		s.mu.Unlock()
		s.registry.OnDisconnect(conn)
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		s.handleMessage(conn, state, msg)
	}
}

func (s *WsServer) handleMessage(conn *websocket.Conn, state *ConnState, raw []byte) {
	var req RequestEnvelope
	if err := json.Unmarshal(raw, &req); err != nil {
		sendReply(conn, ReplyEnvelope{
			ID: "unknown", OK: false, Type: "bad-json",
			Error: &ReplyError{Code: "bad_json", Message: "Invalid JSON"},
		})
		return
	}

	if req.ID == "" || req.Type == "" {
		sendReply(conn, ReplyEnvelope{
			ID: "unknown", OK: false, Type: "bad-request",
			Error: &ReplyError{Code: "bad_request", Message: "Invalid request envelope"},
		})
		return
	}

	switch req.Type {
	case "ping":
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]int64{"ts": time.Now().UnixMilli()}})

	case "list-issues":
		s.handleListIssues(conn, &req)

	case "list-ready":
		s.handleGenericList(conn, &req, "ready-issues")

	case "subscribe-list":
		s.handleSubscribeList(conn, state, &req)

	case "unsubscribe-list":
		s.handleUnsubscribeList(conn, state, &req)

	case "show", "issue-detail":
		s.handleShow(conn, &req)

	case "update-status":
		s.handleUpdateStatus(conn, &req)

	case "edit-text":
		s.handleEditText(conn, &req)

	case "update-priority":
		s.handleBdCommandWithArgs(conn, &req, "priority")

	case "create-issue":
		s.handleCreateIssue(conn, &req)

	case "dep-add":
		s.handleDepCommand(conn, &req, "add")

	case "dep-remove":
		s.handleDepCommand(conn, &req, "remove")

	case "epic-status":
		s.handleBdCommandWithArgs(conn, &req, "epic-status")

	case "update-assignee":
		s.handleBdCommandWithArgs(conn, &req, "assignee")

	case "label-add":
		s.handleLabelCommand(conn, &req, "add")

	case "label-remove":
		s.handleLabelCommand(conn, &req, "remove")

	case "get-comments":
		s.handleGetComments(conn, &req)

	case "add-comment":
		s.handleAddComment(conn, &req)

	case "delete-issue":
		s.handleDeleteIssue(conn, &req)

	case "list-workspaces":
		s.handleListWorkspaces(conn, &req)

	case "set-workspace":
		s.handleSetWorkspace(conn, &req)

	case "get-workspace":
		s.handleGetWorkspace(conn, &req)

	case "get-bd-bin":
		s.handleGetBdBin(conn, &req)

	case "set-bd-bin":
		s.handleSetBdBin(conn, &req)

	case "add-workspace":
		s.handleAddWorkspace(conn, &req)

	case "remove-workspace":
		s.handleRemoveWorkspace(conn, &req)

	default:
		sendReply(conn, ReplyEnvelope{
			ID: req.ID, OK: false, Type: req.Type,
			Error: &ReplyError{Code: "unknown_type", Message: fmt.Sprintf("Unknown message type: %s", req.Type)},
		})
	}
}

func (s *WsServer) handleListIssues(conn *websocket.Conn, req *RequestEnvelope) {
	args := []string{"list", "--json", "--tree=false", "--all"}
	code, parsed, stderr := RunBdJson(args, s.workspace.RootDir)
	if code != 0 {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: stderr}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: parsed})
}

func (s *WsServer) handleGenericList(conn *websocket.Conn, req *RequestEnvelope, subType string) {
	spec := map[string]interface{}{"type": subType}
	result := FetchListForSubscription(spec, s.workspace.RootDir)
	if !result.Ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: result.Error.Code, Message: result.Error.Message}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: result.Items})
}

func (s *WsServer) handleShow(conn *websocket.Conn, req *RequestEnvelope) {
	var id string
	if m, ok := req.Payload.(map[string]interface{}); ok {
		id, _ = m["id"].(string)
	}
	if id == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing id"}})
		return
	}
	code, parsed, stderr := RunBdJson([]string{"show", id, "--json"}, s.workspace.RootDir)
	if code != 0 {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: stderr}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: parsed})
}

func (s *WsServer) handleEditText(conn *websocket.Conn, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	id, _ := m["id"].(string)
	field, _ := m["field"].(string)
	value, _ := m["value"].(string)
	if id == "" || field == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing id or field"}})
		return
	}
	args := []string{"update", id, "--" + field, value}
	result, err := RunBd(args, s.workspace.RootDir)
	if err != nil || result.Code != 0 {
		errMsg := "bd edit failed"
		if err != nil {
			errMsg = err.Error()
		} else if result.Stderr != "" {
			errMsg = result.Stderr
		}
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: errMsg}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{"id": id}})
	s.triggerMutationRefresh()
}

func (s *WsServer) handleUpdateStatus(conn *websocket.Conn, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	id, _ := m["id"].(string)
	status, _ := m["status"].(string)
	if id == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing id"}})
		return
	}
	args := []string{"update", id, "--status", status}
	result, err := RunBd(args, s.workspace.RootDir)
	if err != nil || result.Code != 0 {
		errMsg := "bd update failed"
		if err != nil {
			errMsg = err.Error()
		} else if result.Stderr != "" {
			errMsg = result.Stderr
		}
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: errMsg}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{"id": id}})
	s.triggerMutationRefresh()
}

func (s *WsServer) handleBdCommandWithArgs(conn *websocket.Conn, req *RequestEnvelope, cmdType string) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	id, _ := m["id"].(string)
	if id == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing id"}})
		return
	}

	var args []string
	switch cmdType {
	case "priority":
		priority, _ := m["priority"].(string)
		args = []string{"update", id, "--priority", priority}
	case "assignee":
		assignee, _ := m["assignee"].(string)
		args = []string{"update", id, "--assignee", assignee}
	case "epic-status":
		status, _ := m["status"].(string)
		args = []string{"update", id, "--status", status}
	default:
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Unknown command type"}})
		return
	}

	result, err := RunBd(args, s.workspace.RootDir)
	if err != nil || result.Code != 0 {
		errMsg := "bd command failed"
		if err != nil {
			errMsg = err.Error()
		} else if result.Stderr != "" {
			errMsg = result.Stderr
		}
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: errMsg}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{"id": id}})
	s.triggerMutationRefresh()
}

func (s *WsServer) handleCreateIssue(conn *websocket.Conn, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	title, _ := m["title"].(string)
	if title == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing title"}})
		return
	}
	args := []string{"create", title}
	if body, _ := m["body"].(string); body != "" {
		args = append(args, "--body", body)
	}
	if issueType, _ := m["issue_type"].(string); issueType != "" {
		args = append(args, "--type", issueType)
	}
	if priority, _ := m["priority"].(string); priority != "" {
		args = append(args, "--priority", priority)
	}
	if labels, _ := m["labels"].(string); labels != "" {
		args = append(args, "--labels", labels)
	}

	result, err := RunBd(args, s.workspace.RootDir)
	if err != nil || result.Code != 0 {
		errMsg := "bd create failed"
		if err != nil {
			errMsg = err.Error()
		} else if result.Stderr != "" {
			errMsg = result.Stderr
		}
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: errMsg}})
		return
	}
	code, parsed, _ := RunBdJson([]string{"list", "--json", "--tree=false", "--limit", "1"}, s.workspace.RootDir)
	if code == 0 && parsed != nil {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: parsed})
	} else {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{"title": title}})
	}
	s.triggerMutationRefresh()
}

func (s *WsServer) handleDepCommand(conn *websocket.Conn, req *RequestEnvelope, action string) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	id, _ := m["id"].(string)
	depID, _ := m["dep_id"].(string)
	if id == "" || depID == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing id or dep_id"}})
		return
	}
	args := []string{"dep", action, id, depID}
	result, err := RunBd(args, s.workspace.RootDir)
	if err != nil || result.Code != 0 {
		errMsg := "bd dep failed"
		if err != nil {
			errMsg = err.Error()
		} else if result.Stderr != "" {
			errMsg = result.Stderr
		}
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: errMsg}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{"id": id}})
	s.triggerMutationRefresh()
}

func (s *WsServer) handleLabelCommand(conn *websocket.Conn, req *RequestEnvelope, action string) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	id, _ := m["id"].(string)
	label, _ := m["label"].(string)
	if id == "" || label == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing id or label"}})
		return
	}
	args := []string{"label", action, id, label}
	result, err := RunBd(args, s.workspace.RootDir)
	if err != nil || result.Code != 0 {
		errMsg := "bd label failed"
		if err != nil {
			errMsg = err.Error()
		} else if result.Stderr != "" {
			errMsg = result.Stderr
		}
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: errMsg}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{"id": id}})
	s.triggerMutationRefresh()
}

func (s *WsServer) handleGetComments(conn *websocket.Conn, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	id, _ := m["id"].(string)
	if id == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing id"}})
		return
	}
	code, parsed, stderr := RunBdJson([]string{"comments", id, "--json"}, s.workspace.RootDir)
	if code != 0 {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: stderr}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: parsed})
}

func (s *WsServer) handleAddComment(conn *websocket.Conn, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	id, _ := m["id"].(string)
	body, _ := m["body"].(string)
	if id == "" || body == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing id or body"}})
		return
	}
	args := []string{"comment", id, body}
	result, err := RunBd(args, s.workspace.RootDir)
	if err != nil || result.Code != 0 {
		errMsg := "bd comment failed"
		if err != nil {
			errMsg = err.Error()
		} else if result.Stderr != "" {
			errMsg = result.Stderr
		}
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: errMsg}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{"id": id}})
	s.triggerMutationRefresh()
}

func (s *WsServer) handleDeleteIssue(conn *websocket.Conn, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	id, _ := m["id"].(string)
	if id == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing id"}})
		return
	}
	args := []string{"delete", "-f", id}
	result, err := RunBd(args, s.workspace.RootDir)
	if err != nil || result.Code != 0 {
		errMsg := "bd delete failed"
		if err != nil {
			errMsg = err.Error()
		} else if result.Stderr != "" {
			errMsg = result.Stderr
		}
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: errMsg}})
		return
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{"id": id}})
	s.triggerMutationRefresh()
}

func (s *WsServer) handleListWorkspaces(conn *websocket.Conn, req *RequestEnvelope) {
	workspaces := GetAvailableWorkspaces()
	payload := map[string]interface{}{
		"workspaces": workspaces,
		"current": map[string]string{
			"root_dir": s.workspace.RootDir,
			"db_path":  s.workspace.DbPath,
		},
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: payload})
}

func (s *WsServer) handleSetWorkspace(conn *websocket.Conn, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	path, _ := m["path"].(string)
	if path == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing path"}})
		return
	}

	absPath, _ := filepath.Abs(path)
	newDb := ResolveWorkspaceDatabase(absPath)
	oldPath := s.workspace.DbPath

	s.workspace.RootDir = absPath
	s.workspace.DbPath = newDb.Path

	changed := newDb.Path != oldPath
	if changed && s.watcher != nil {
		s.watcher.Rebind(absPath)
		s.registry.Clear()
		s.broadcast("workspace-changed", map[string]string{
			"root_dir": absPath,
			"db_path":  newDb.Path,
		})
		s.ScheduleListRefresh()
	}

	payload := map[string]interface{}{
		"changed": changed,
		"workspace": map[string]string{
			"root_dir": absPath,
			"db_path":  newDb.Path,
		},
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: payload})
}

func (s *WsServer) handleGetWorkspace(conn *websocket.Conn, req *RequestEnvelope) {
	payload := map[string]string{
		"root_dir": s.workspace.RootDir,
		"db_path":  s.workspace.DbPath,
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: payload})
}

func (s *WsServer) handleGetBdBin(conn *websocket.Conn, req *RequestEnvelope) {
	bin := GetBdBin()
	result, err := RunBd([]string{"--version"}, s.workspace.RootDir)
	version := ""
	if err == nil && result.Code == 0 {
		version = result.Stdout
	}
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{
		"path":    bin,
		"version": version,
	}})
}

func (s *WsServer) handleSetBdBin(conn *websocket.Conn, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	path, _ := m["path"].(string)
	if path == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing path"}})
		return
	}

	SetBdBin(path)

	newBin := GetBdBin()
	result, err := RunBd([]string{"--version"}, s.workspace.RootDir)
	version := ""
	if err == nil && result.Code == 0 {
		version = result.Stdout
	}

	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{
		"path":    newBin,
		"version": version,
	}})
}

func (s *WsServer) handleAddWorkspace(conn *websocket.Conn, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	path, _ := m["path"].(string)
	if path == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing path"}})
		return
	}

	absPath, _ := filepath.Abs(path)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "not_found", Message: "Directory does not exist"}})
		return
	}

	metadataFile := filepath.Join(absPath, ".beads", "metadata.json")
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "not_beads", Message: "Not a beads project: .beads/metadata.json not found"}})
		return
	}

	result, err := RunBd([]string{"list", "--json", "--tree=false"}, absPath)
	if err != nil || result.Code != 0 {
		errMsg := "Not a valid beads project"
		if result != nil && result.Stderr != "" {
			errMsg = result.Stderr
		}
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bd_error", Message: errMsg}})
		return
	}

	info := AddWorkspace(absPath)
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{
		"path":     info.Path,
		"database": info.Database,
	}})
}

func (s *WsServer) handleRemoveWorkspace(conn *websocket.Conn, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}
	path, _ := m["path"].(string)
	if path == "" {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing path"}})
		return
	}

	absPath, _ := filepath.Abs(path)
	RemoveWorkspace(absPath)
	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{
		"removed": absPath,
	}})
}

func (s *WsServer) handleSubscribeList(conn *websocket.Conn, state *ConnState, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}

	clientID, _ := m["id"].(string)
	spec, _ := m["spec"].(map[string]interface{})
	if clientID == "" || spec == nil {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Missing id or spec"}})
		return
	}

	key := s.registry.Attach(spec, conn)

	state.mu.Lock()
	state.ListSubs[clientID] = SubSpec{Key: key, Spec: spec}
	state.mu.Unlock()

	result := FetchListForSubscription(spec, s.workspace.RootDir)
	if result.Ok {
		s.emitSubscriptionSnapshot(conn, clientID, key, result.Items)
	} else {
		s.emitSubscriptionSnapshot(conn, clientID, key, []map[string]interface{}{})
	}

	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{"subscribed": clientID}})
}

func (s *WsServer) handleUnsubscribeList(conn *websocket.Conn, state *ConnState, req *RequestEnvelope) {
	m, ok := req.Payload.(map[string]interface{})
	if !ok {
		sendReply(conn, ReplyEnvelope{ID: req.ID, OK: false, Type: req.Type, Error: &ReplyError{Code: "bad_request", Message: "Invalid payload"}})
		return
	}

	clientID, _ := m["id"].(string)
	state.mu.Lock()
	sub, exists := state.ListSubs[clientID]
	delete(state.ListSubs, clientID)
	state.mu.Unlock()

	if exists {
		s.registry.Detach(sub.Spec, conn)
	}

	sendReply(conn, ReplyEnvelope{ID: req.ID, OK: true, Type: req.Type, Payload: map[string]string{"unsubscribed": clientID}})
}

func (s *WsServer) ScheduleListRefresh() {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()

	if s.refreshTimer != nil {
		s.refreshTimer.Stop()
	}
	s.refreshTimer = time.AfterFunc(75*time.Millisecond, func() {
		s.refreshAllActiveListSubscriptions()
	})
}

func (s *WsServer) triggerMutationRefresh() {
	time.AfterFunc(100*time.Millisecond, func() {
		s.ScheduleListRefresh()
	})
}

func (s *WsServer) refreshAllActiveListSubscriptions() {
	specs := s.collectActiveListSpecs()
	for _, spec := range specs {
		s.refreshAndPublish(spec)
	}
}

func (s *WsServer) collectActiveListSpecs() []map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	seen := make(map[string]bool)
	var specs []map[string]interface{}

	for _, state := range s.connections {
		state.mu.Lock()
		for _, sub := range state.ListSubs {
			if !seen[sub.Key] {
				seen[sub.Key] = true
				specs = append(specs, sub.Spec)
			}
		}
		state.mu.Unlock()
	}
	return specs
}

func (s *WsServer) refreshAndPublish(spec map[string]interface{}) {
	key := KeyOf(spec)
	result := FetchListForSubscription(spec, s.workspace.RootDir)
	if !result.Ok {
		return
	}

	delta := s.registry.ApplyItems(key, result.Items)
	entry := s.registry.Get(key)
	if entry == nil {
		return
	}

	byID := make(map[string]map[string]interface{})
	for _, it := range result.Items {
		if id, ok := it["id"].(string); ok {
			byID[id] = it
		}
	}

	entry.mu.Lock()
	subscribers := make([]*websocket.Conn, 0, len(entry.Subscribers))
	for ws := range entry.Subscribers {
		subscribers = append(subscribers, ws)
	}
	entry.mu.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ws := range subscribers {
		if ws == nil {
			continue
		}
		state, ok := s.connections[ws]
		if !ok {
			continue
		}
		state.mu.Lock()
		for clientID, sub := range state.ListSubs {
			if sub.Key != key {
				continue
			}
			if len(delta.Added) == 0 && len(delta.Updated) == 0 && len(delta.Removed) == 0 {
				continue
			}
			for _, id := range delta.Added {
				if issue, ok := byID[id]; ok {
					s.emitSubscriptionUpsert(ws, clientID, key, issue)
				}
			}
			for _, id := range delta.Updated {
				if issue, ok := byID[id]; ok {
					s.emitSubscriptionUpsert(ws, clientID, key, issue)
				}
			}
			for _, id := range delta.Removed {
				s.emitSubscriptionDelete(ws, clientID, key, id)
			}
		}
		state.mu.Unlock()
	}
}

func (s *WsServer) emitSubscriptionSnapshot(conn *websocket.Conn, clientID, key string, issues []map[string]interface{}) {
	payload := map[string]interface{}{
		"type":     "snapshot",
		"id":       clientID,
		"revision": 1,
		"issues":   issues,
	}
	msg, _ := json.Marshal(ReplyEnvelope{
		ID:      fmt.Sprintf("evt-%d", time.Now().UnixMilli()),
		OK:      true,
		Type:    "snapshot",
		Payload: payload,
	})
	conn.WriteMessage(websocket.TextMessage, msg)
}

func (s *WsServer) emitSubscriptionUpsert(conn *websocket.Conn, clientID, key string, issue map[string]interface{}) {
	payload := map[string]interface{}{
		"type":     "upsert",
		"id":       clientID,
		"revision": 1,
		"issue":    issue,
	}
	msg, _ := json.Marshal(ReplyEnvelope{
		ID:      fmt.Sprintf("evt-%d", time.Now().UnixMilli()),
		OK:      true,
		Type:    "upsert",
		Payload: payload,
	})
	conn.WriteMessage(websocket.TextMessage, msg)
}

func (s *WsServer) emitSubscriptionDelete(conn *websocket.Conn, clientID, key, issueID string) {
	payload := map[string]interface{}{
		"type":     "delete",
		"id":       clientID,
		"revision": 1,
		"issue_id": issueID,
	}
	msg, _ := json.Marshal(ReplyEnvelope{
		ID:      fmt.Sprintf("evt-%d", time.Now().UnixMilli()),
		OK:      true,
		Type:    "delete",
		Payload: payload,
	})
	conn.WriteMessage(websocket.TextMessage, msg)
}

func (s *WsServer) broadcast(msgType string, payload interface{}) {
	msg, _ := json.Marshal(ReplyEnvelope{
		ID:      fmt.Sprintf("evt-%d", time.Now().UnixMilli()),
		OK:      true,
		Type:    msgType,
		Payload: payload,
	})

	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.connections {
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func sendReply(conn *websocket.Conn, reply ReplyEnvelope) {
	msg, err := json.Marshal(reply)
	if err != nil {
		return
	}
	conn.WriteMessage(websocket.TextMessage, msg)
}
