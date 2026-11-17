package node

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/nwverse/persona/server/internal/proto"
	"github.com/nwverse/persona/server/internal/tokenuse"
)

// Runtime usage is a per-request delta, not a replayable local transcript. Remember
// request IDs across runtime reconnects on this node connection; a node restart loses
// this memory and in-flight reports. There is deliberately no historical backfill.
type runtimeUsageReports struct {
	mu       sync.Mutex
	seen     map[runtimeUsageRequest]time.Time
	prunedAt time.Time
}
type runtimeUsageRequest struct{ agent, session, request string }

type runtimeUsageReport struct {
	SessionID  string    `json:"session_id"`
	TS         time.Time `json:"ts"`
	Model      string    `json:"model"`
	RequestID  string    `json:"request_id"`
	Input      int64     `json:"input"`
	Output     int64     `json:"output"`
	CacheRead  int64     `json:"cache_read"`
	CacheWrite int64     `json:"cache_write"`
	Reasoning  int64     `json:"reasoning"`
}

func (n *Node) runtimeUsageReport(h *runtimeConn, params json.RawMessage, now time.Time) {
	var p runtimeUsageReport
	if json.Unmarshal(params, &p) != nil || n.peer == nil || n.runtimes == nil || !n.runtimes.owns(h, p.SessionID) {
		return
	}
	if p.TS.Before(now.Add(-usageWindow)) || p.TS.After(now.Add(time.Minute)) || p.Model == "" || len(p.Model) > 256 || len(p.RequestID) > 512 {
		return
	}
	// The wire uses JS-safe nonnegative integer counts. Reasoning is already in output.
	total := int64(0)
	for _, v := range []int64{p.Input, p.Output, p.CacheRead, p.CacheWrite} {
		if v < 0 || v > 1<<53-1-total {
			return
		}
		total += v
	}
	if total == 0 || p.Reasoning < 0 || p.Reasoning > p.Output {
		return
	}
	// Attribute from the owned row, never from fields supplied with the event. An
	// engine can host several agents (Paseo); its hello's agent alone is insufficient.
	agent := ""
	h.mu.Lock()
	for _, s := range h.sessions {
		if s.SessionID == p.SessionID {
			agent = s.Agent
			break
		}
	}
	h.mu.Unlock()
	if agent == "" || tokenuse.FileMetered(agent) {
		return
	}
	st, _ := n.state.get(p.SessionID)
	bucket := proto.UsageBucket{Minute: p.TS.Unix() / 60 * 60, Model: p.Model, Launcher: st.Launcher, Agent: agent,
		Input: p.Input, Output: p.Output, CacheRead: p.CacheRead, CacheWrite: p.CacheWrite, Requests: 1}
	u := &n.runtimeUsage
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.seen == nil {
		u.seen = map[runtimeUsageRequest]time.Time{}
	}
	if u.prunedAt.IsZero() || now.Sub(u.prunedAt) >= time.Minute {
		for k, t := range u.seen {
			if t.Before(now.Add(-usageWindow)) {
				delete(u.seen, k)
			}
		}
		u.prunedAt = now
	}
	key := runtimeUsageRequest{agent, p.SessionID, p.RequestID}
	if p.RequestID != "" {
		if _, ok := u.seen[key]; ok {
			return
		}
		if len(u.seen) >= 65536 {
			return
		}
	}
	if err := n.peer.Notify("usage", map[string]any{"add": true, "buckets": []proto.UsageBucket{bucket}}); err == nil && p.RequestID != "" {
		u.seen[key] = p.TS
	}
}
