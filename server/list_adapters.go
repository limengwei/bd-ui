package server

import (
	"fmt"
	"time"
)

func MapSubscriptionToBdArgs(spec map[string]interface{}) ([]string, error) {
	t, _ := spec["type"].(string)
	switch t {
	case "all-issues":
		return []string{"list", "--json", "--tree=false", "--all"}, nil
	case "epics":
		return []string{"epic", "status", "--json"}, nil
	case "blocked-issues":
		return []string{"blocked", "--json"}, nil
	case "ready-issues":
		return []string{"ready", "--limit", "1000", "--json"}, nil
	case "in-progress-issues":
		return []string{"list", "--json", "--tree=false", "--status", "in_progress"}, nil
	case "closed-issues":
		return []string{"list", "--json", "--tree=false", "--status", "closed", "--limit", "1000"}, nil
	case "issue-detail":
		params, _ := spec["params"].(map[string]interface{})
		id, _ := params["id"].(string)
		if id == "" {
			return nil, fmt.Errorf("Missing param: params.id")
		}
		return []string{"show", id, "--json"}, nil
	default:
		return nil, fmt.Errorf("Unknown subscription type: %s", t)
	}
}

type FetchListResult struct {
	Ok    bool                     `json:"ok"`
	Items []map[string]interface{} `json:"items,omitempty"`
	Error *FetchListError          `json:"error,omitempty"`
}

type FetchListError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func FetchListForSubscription(spec map[string]interface{}, cwd string) *FetchListResult {
	args, err := MapSubscriptionToBdArgs(spec)
	if err != nil {
		return &FetchListResult{
			Ok: false,
			Error: &FetchListError{
				Code:    "bad_request",
				Message: err.Error(),
			},
		}
	}

	code, parsed, stderr := RunBdJson(args, cwd)
	if code != 0 || parsed == nil {
		errMsg := stderr
		if errMsg == "" {
			errMsg = "bd failed"
		}
		return &FetchListResult{
			Ok: false,
			Error: &FetchListError{
				Code:    "bd_error",
				Message: errMsg,
			},
		}
	}

	t, _ := spec["type"].(string)
	items := normalizeIssueList(parsed, t)
	return &FetchListResult{Ok: true, Items: items}
}

func normalizeIssueList(raw interface{}, specType string) []map[string]interface{} {
	var items []map[string]interface{}

	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, m)
			}
		}
	case map[string]interface{}:
		items = append(items, v)
	}

	if specType == "epics" {
		var filtered []map[string]interface{}
		for _, it := range items {
			if epic, ok := it["epic"].(map[string]interface{}); ok {
				flat := map[string]interface{}{
					"id":                 toString(epic["id"]),
					"title":              epic["title"],
					"status":             epic["status"],
					"issue_type":         epic["issue_type"],
					"created_at":         epic["created_at"],
					"updated_at":         epic["updated_at"],
					"closed_at":          epic["closed_at"],
					"total_children":     it["total_children"],
					"closed_children":    it["closed_children"],
					"eligible_for_close": it["eligible_for_close"],
				}
				if flat["issue_type"] == nil {
					flat["issue_type"] = "epic"
				}
				filtered = append(filtered, flat)
			} else {
				filtered = append(filtered, it)
			}
		}
		var result []map[string]interface{}
		for _, it := range filtered {
			status, _ := it["status"].(string)
			if status == "tombstone" {
				continue
			}
			if it["deleted_at"] != nil {
				continue
			}
			result = append(result, it)
		}
		items = result
	}

	var result []map[string]interface{}
	for _, it := range items {
		id, _ := it["id"].(string)
		if id == "" {
			continue
		}
		it["id"] = id
		if _, ok := it["created_at"]; ok {
			it["created_at"] = toTimestamp(it["created_at"])
		}
		it["updated_at"] = toTimestamp(it["updated_at"])
		if it["closed_at"] != nil {
			it["closed_at"] = toTimestamp(it["closed_at"])
		}
		if deps, ok := it["dependencies"].([]interface{}); ok {
			var depIDs []string
			for _, d := range deps {
				if dm, ok := d.(map[string]interface{}); ok {
					if depID, ok := dm["depends_on_id"].(string); ok && depID != "" {
						depIDs = append(depIDs, depID)
					}
				}
			}
			if len(depIDs) > 0 {
				it["dep_ids"] = depIDs
			}
			delete(it, "dependencies")
		}
		result = append(result, it)
	}

	enrichEpicProgress(result)

	if len(result) == 0 {
		return []map[string]interface{}{}
	}
	return result
}

func enrichEpicProgress(items []map[string]interface{}) {
	childrenMap := make(map[string][]map[string]interface{})
	for _, it := range items {
		parent, _ := it["parent"].(string)
		if parent != "" {
			childrenMap[parent] = append(childrenMap[parent], it)
		}
	}
	if len(childrenMap) == 0 {
		return
	}
	for _, it := range items {
		issueType, _ := it["issue_type"].(string)
		if issueType != "epic" {
			continue
		}
		id, _ := it["id"].(string)
		children := childrenMap[id]
		total := len(children)
		closed := 0
		for _, c := range children {
			status, _ := c["status"].(string)
			if status == "closed" {
				closed++
			}
		}
		it["total_children"] = total
		it["closed_children"] = closed
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func toTimestamp(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case string:
		if t, err := time.Parse(time.RFC3339, n); err == nil {
			return float64(t.UnixMilli())
		}
		return 0
	default:
		return 0
	}
}
