package hub

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"github.com/nwverse/persona/server/internal/extwire"
	"github.com/nwverse/persona/server/internal/proto"
)

func (h *Hub) scopedQuestion(ix proto.Interaction) bool {
	return ix.Kind == proto.IxQuestion && ix.ToolName != proto.SecretHandoffTool && !h.hiddenFromExtensions(ix)
}

func (h *Hub) scopedPublicSession(machine, session string) bool {
	m := h.localMachine(machine)
	if m == nil {
		return false
	}
	s, ok := m.session(session)
	return ok && !healthSession(&s)
}

// scopedDeviceRoute uses the same registered route pattern as extension
// authorization. A scoped device can receive only scopes for Agent integration.
func (h *Hub) scopedDeviceRoute(r *http.Request) string {
	if h.apiMux == nil {
		return ""
	}
	_, pattern := h.apiMux.Handler(r)
	switch pattern {
	case "GET /api/v1/interactions/history", "POST /api/v1/interactions/{id}/reanswer":
		// Stored answers have no pending card to prove their visibility or kind here.
		return ""
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

// Scoped sockets can follow conversations and check connectivity.
func scopedSocketCommand(command string) bool {
	switch command {
	case "subscribe", "unsubscribe", "ping":
		return true
	default:
		return false
	}
}

// scopedSocketEvent recognizes frames supported by a scoped chat socket.
func scopedSocketEvent(event string) bool {
	switch event {
	case "hello", "pong", "sessions", "items", "delta", "live":
		return true
	default:
		return false
	}
}

// Scoped sockets never replay the owner event log. Project the initial session
// snapshot and the connection's subscribed chat frames into their public shape.
func (c *client) scopedSocketFrame(data []byte) []byte {
	var frame map[string]json.RawMessage
	if json.Unmarshal(data, &frame) != nil {
		return nil
	}
	var kind string
	if json.Unmarshal(frame["type"], &kind) != nil || !scopedSocketEvent(kind) {
		return nil
	}
	out := map[string]any{"type": kind}
	copyFields := func(keys ...string) {
		for _, key := range keys {
			if value, exists := frame[key]; exists {
				out[key] = value
			}
		}
	}
	switch kind {
	case "hello":
		copyFields("server_version", "protocol", "client_id")
	case "pong":
		copyFields("id")
	case "sessions":
		var rows []proto.Session
		if json.Unmarshal(frame["sessions"], &rows) != nil {
			return nil
		}
		out["sessions"] = slices.DeleteFunc(rows, func(s proto.Session) bool { return healthSession(&s) })
		copyFields("machine_id")
	default:
		var mid, sid string
		if json.Unmarshal(frame["machine_id"], &mid) != nil || json.Unmarshal(frame["session_id"], &sid) != nil ||
			!c.hub.scopedPublicSession(mid, sid) {
			return nil
		}
		c.mu.Lock()
		subscribed := c.subs[[2]string{mid, sid}]
		c.mu.Unlock()
		if !subscribed {
			return nil
		}
		copyFields("machine_id", "session_id")
		switch kind {
		case "items":
			var items []proto.Item
			if json.Unmarshal(frame["items"], &items) != nil {
				return nil
			}
			out["items"] = items
			copyFields("history", "from_seq", "to_seq", "backlog", "more_after")
		case "delta":
			copyFields("text")
		case "live":
			var live *proto.SessionLive
			if json.Unmarshal(frame["live"], &live) != nil {
				return nil
			}
			out["live"] = live
		}
	}
	projected, _ := json.Marshal(out)
	return projected
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
