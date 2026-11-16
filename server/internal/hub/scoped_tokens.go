package hub

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"github.com/nwverse/persona/server/internal/extwire"
)

// scopedDeviceRoute uses the same registered route pattern as extension
// authorization. A scoped device can receive only scopes for Agent integration.
func (h *Hub) scopedDeviceRoute(r *http.Request) string {
	if h.apiMux == nil {
		return ""
	}
	_, pattern := h.apiMux.Handler(r)
	switch pattern {
	case "GET /api/v1/ws",
		"GET /api/v1/machines/{mid}/launchers",
		"GET /api/v1/machines/{mid}/launchers/{name}/models":
		return extwire.ScopeAgentsRead
	}
	return extwire.RouteScopes[pattern]
}

func deviceAgentScope(scope string) bool {
	switch scope {
	case extwire.ScopeAgentsRead, extwire.ScopeAgentsControl, extwire.ScopeAgentsPermissionMode:
		return true
	default:
		return false
	}
}

func scopedPermissionDenied(w http.ResponseWriter, scope string) {
	body := errorBody(http.StatusForbidden, "permission_denied", errors.New("this credential cannot access this part of the hub API"))
	if scope != "" {
		body["scope"] = scope
	}
	writeErrorBody(w, http.StatusForbidden, body)
}

func (h *Hub) authorizeScopedDevice(w http.ResponseWriter, r *http.Request, p principal) bool {
	scope := h.scopedDeviceRoute(r)
	if deviceAgentScope(scope) && slices.Contains(p.Scopes, scope) {
		return true
	}
	scopedPermissionDenied(w, scope)
	return false
}

// scopedSocketEvent fails closed for every push outside the chat stream.
func scopedSocketEvent(data []byte) bool {
	var event struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(data, &event) != nil {
		return false
	}
	switch event.Type {
	case "hello", "pong", "sessions", "sessions_delta", "sessions_list",
		"items", "delta", "live", "message_send_state", "conversation_deleted", "error":
		return true
	default:
		return false
	}
}

func validDeviceScopes(scopes []string) bool {
	if len(scopes) == 0 {
		return false
	}
	for index, scope := range scopes {
		if !deviceAgentScope(scope) || slices.Contains(scopes[:index], scope) {
			return false
		}
	}
	return true
}
