package server

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type ItemMeta struct {
	UpdatedAt float64 `json:"updated_at"`
	ClosedAt  *float64 `json:"closed_at"`
}

type SubEntry struct {
	mu          sync.Mutex
	ItemsById   map[string]ItemMeta
	Subscribers map[*websocket.Conn]bool
}

type SubRegistry struct {
	mu      sync.Mutex
	entries map[string]*SubEntry
}

func NewSubRegistry() *SubRegistry {
	return &SubRegistry{
		entries: make(map[string]*SubEntry),
	}
}

func KeyOf(spec map[string]interface{}) string {
	t := ""
	if v, ok := spec["type"].(string); ok {
		t = strings.TrimSpace(v)
	}

	params, _ := spec["params"].(map[string]interface{})
	if len(params) == 0 {
		return t
	}

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	vals := url.Values{}
	for _, k := range keys {
		vals.Set(k, fmt.Sprintf("%v", params[k]))
	}
	enc := vals.Encode()
	if enc != "" {
		return t + "?" + enc
	}
	return t
}

func (r *SubRegistry) Get(key string) *SubEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.entries[key]
}

func (r *SubRegistry) Ensure(spec map[string]interface{}) (string, *SubEntry) {
	key := KeyOf(spec)
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.entries[key]
	if !ok {
		entry = &SubEntry{
			ItemsById:   make(map[string]ItemMeta),
			Subscribers: make(map[*websocket.Conn]bool),
		}
		r.entries[key] = entry
	}
	return key, entry
}

func (r *SubRegistry) Attach(spec map[string]interface{}, ws *websocket.Conn) string {
	key, entry := r.Ensure(spec)
	entry.mu.Lock()
	entry.Subscribers[ws] = true
	entry.mu.Unlock()
	return key
}

func (r *SubRegistry) Detach(spec map[string]interface{}, ws *websocket.Conn) bool {
	key := KeyOf(spec)
	r.mu.Lock()
	entry, ok := r.entries[key]
	r.mu.Unlock()
	if !ok {
		return false
	}
	entry.mu.Lock()
	defer entry.mu.Unlock()
	delete(entry.Subscribers, ws)
	return true
}

func (r *SubRegistry) OnDisconnect(ws *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var empties []string
	for key, entry := range r.entries {
		entry.mu.Lock()
		delete(entry.Subscribers, ws)
		if len(entry.Subscribers) == 0 {
			empties = append(empties, key)
		}
		entry.mu.Unlock()
	}
	for _, key := range empties {
		delete(r.entries, key)
	}
}

type Delta struct {
	Added   []string `json:"added"`
	Updated []string `json:"updated"`
	Removed []string `json:"removed"`
}

func ComputeDelta(prev, next map[string]ItemMeta) Delta {
	var d Delta
	for id, meta := range next {
		p, ok := prev[id]
		if !ok {
			d.Added = append(d.Added, id)
			continue
		}
		if p.UpdatedAt != meta.UpdatedAt || !closedAtEqual(p.ClosedAt, meta.ClosedAt) {
			d.Updated = append(d.Updated, id)
		}
	}
	for id := range prev {
		if _, ok := next[id]; !ok {
			d.Removed = append(d.Removed, id)
		}
	}
	return d
}

func closedAtEqual(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ToItemsMap(items []map[string]interface{}) map[string]ItemMeta {
	m := make(map[string]ItemMeta)
	for _, it := range items {
		id, _ := it["id"].(string)
		if id == "" {
			continue
		}
		var updatedAt float64
		if v, ok := toFloat(it["updated_at"]); ok {
			updatedAt = v
		}
		var closedAt *float64
		if v, ok := it["closed_at"]; ok && v != nil {
			if f, ok := toFloat(v); ok {
				closedAt = &f
			}
		}
		m[id] = ItemMeta{UpdatedAt: updatedAt, ClosedAt: closedAt}
	}
	return m
}

func (r *SubRegistry) ApplyItems(key string, items []map[string]interface{}) Delta {
	nextMap := ToItemsMap(items)
	r.mu.Lock()
	entry, ok := r.entries[key]
	if !ok {
		entry = &SubEntry{
			ItemsById:   make(map[string]ItemMeta),
			Subscribers: make(map[*websocket.Conn]bool),
		}
		r.entries[key] = entry
	}
	r.mu.Unlock()

	entry.mu.Lock()
	defer entry.mu.Unlock()
	delta := ComputeDelta(entry.ItemsById, nextMap)
	entry.ItemsById = make(map[string]ItemMeta)
	for k, v := range nextMap {
		entry.ItemsById[k] = v
	}
	return delta
}

func (r *SubRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = make(map[string]*SubEntry)
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		return 0, false
	default:
		return 0, false
	}
}
