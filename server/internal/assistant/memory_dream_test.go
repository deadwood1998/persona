package assistant

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/nwverse/persona/server/internal/proto"
)

func TestMemoryDreamUsesOnlyNewUserText(t *testing.T) {
	items := []proto.Item{
		{Seq: 1, Role: "user", Kind: proto.KindText, Text: "older"},
		{Seq: 2, Role: "user", Kind: proto.KindText, Text: "first"},
		{Seq: 3, Role: "user", Kind: proto.KindText, Text: "automation", Source: &proto.MessageSource{Kind: proto.SourceConversation}},
		{Seq: 4, Role: "assistant", Kind: proto.KindText, Text: "reply"},
		{Seq: 5, Role: "user", Kind: proto.KindText, Text: "tool result", ToolUseID: "tool"},
		{Seq: 6, Role: "user", Kind: proto.KindText, Text: "second"},
		{Seq: 7, Role: "user", Kind: proto.KindText, Text: "truncated", FullSize: 42},
	}
	messages, seq := memoryDreamMessages(items, 1)
	if !reflect.DeepEqual(messages, []string{"first", "second"}) || seq != 6 {
		t.Fatalf("messages = %q, seq = %d", messages, seq)
	}
}

type dreamTestHub struct {
	*fakeHub
	writes []struct {
		Cwd     string
		Content string
	}
}

func (h *dreamTestHub) DeviceCall(ctx context.Context, p DeviceCall) (json.RawMessage, error) {
	if p.Method != DeviceAssistantMemory {
		return h.fakeHub.DeviceCall(ctx, p)
	}
	var op struct {
		Op      string
		Cwd     string
		Content string
	}
	if err := json.Unmarshal(p.Params, &op); err != nil {
		return nil, err
	}
	if op.Op == "dream-read" {
		return json.RawMessage(`{"note":null,"index":""}`), nil
	}
	h.writes = append(h.writes, struct {
		Cwd     string
		Content string
	}{op.Cwd, op.Content})
	return json.RawMessage(`{"ok":true}`), nil
}

func TestMemoryDreamProcessesIdleMainOnce(t *testing.T) {
	ctx := context.Background()
	f := newFakeHub().withMachine()
	f.devices = []Device{{ID: "server", Online: true, Home: "/home/example"}}
	f.sessions["server"] = []proto.Session{{SessionID: "main", Assistant: true, Status: proto.StatusDone}}
	f.items = []proto.Item{{Seq: 7, Role: "user", Kind: proto.KindText, Text: "I always prefer tea."}}
	f.judge = `{"body":"Prefers tea."}`
	h := &dreamTestHub{fakeHub: f}
	st := &memStore{}
	s := withSettings(t, f, New(Options{Hub: h, Store: func() Store { return st }, Logf: testLogf(t)}))
	setSettings(t, s, func(v *Settings) {
		v.Main, v.MainMachineID, v.MachineID = "main", "server", "server"
	})
	s.memoryDream(ctx)
	if f.judged != 0 {
		t.Fatal("disabled dreaming called the model")
	}
	f.mu.Lock()
	f.hello.Assistant.MemoryDreaming = true
	f.mu.Unlock()
	s.memoryDream(ctx)
	s.memoryDream(ctx)
	calls := f.judged
	if calls != 1 || len(h.writes) != 1 || h.writes[0].Content != "Prefers tea." || h.writes[0].Cwd == "" {
		t.Fatalf("model calls = %d, writes = %+v", calls, h.writes)
	}
}
