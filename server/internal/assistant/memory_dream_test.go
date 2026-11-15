package assistant

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/nwverse/persona/server/internal/proto"
)

func TestMemoryDreamUsesOnlyNewUserText(t *testing.T) {
	const legacyWake = "⏰（Persona 叫醒你，第 1 次）你上一轮停下时说在等「检查完成」" // i18n-ignore: legacy injected prompt
	items := []proto.Item{
		{Seq: 1, Role: "user", Kind: proto.KindText, Text: "older"},
		{Seq: 2, Role: "user", Kind: proto.KindText, Text: "first"},
		{Seq: 3, Role: "user", Kind: proto.KindText, Text: "automation", Source: &proto.MessageSource{Kind: proto.SourceConversation}},
		{Seq: 4, Role: "assistant", Kind: proto.KindText, Text: "reply"},
		{Seq: 5, Role: "user", Kind: proto.KindText, Text: "tool result", ToolUseID: "tool"},
		{Seq: 6, Role: "user", Kind: proto.KindText, Text: "second"},
		{Seq: 7, Role: "user", Kind: proto.KindText, Text: "truncated", FullSize: 42},
		{Seq: 8, Role: "user", Kind: proto.KindText, Text: legacyWake, ClientMsgID: "typed-by-user"},
		{Seq: 9, Role: "user", Kind: proto.KindText, Text: "new wake-up text", Injected: &proto.Injected{Kind: proto.InjectedFollowup, State: "wake"}},
		{Seq: 10, Role: "user", Kind: proto.KindText, Text: legacyWake},
		{Seq: 11, Role: "user", Kind: proto.KindText, Text: "relayed from another chat", From: "worker"},
	}
	messages, seq := memoryDreamMessages(items, 1)
	if !reflect.DeepEqual(messages, []string{"first", "second", legacyWake}) || seq != 8 {
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
	f.items = []proto.Item{{Seq: 6, Role: "user", Kind: proto.KindText, Text: "Automated reminder", Injected: &proto.Injected{Kind: proto.InjectedFollowup, State: "wake"}}}
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
	if f.judged != 0 || len(h.writes) != 0 {
		t.Fatal("a transcript with only injected text triggered memory dreaming")
	}
	f.items = append(f.items, proto.Item{Seq: 7, Role: "user", Kind: proto.KindText, Text: "I always prefer tea."})
	s.memoryDream(ctx)
	s.memoryDream(ctx)
	calls := f.judged
	if calls != 1 || len(h.writes) != 1 || h.writes[0].Content != "Prefers tea." || h.writes[0].Cwd == "" {
		t.Fatalf("model calls = %d, writes = %+v", calls, h.writes)
	}
}

func TestMemoryDreamTracksTranscriptHistory(t *testing.T) {
	ctx := t.Context()
	f := newFakeHub().withMachine()
	f.devices = []Device{{ID: "server", Online: true, Home: "/home/example"}}
	f.sessions["server"] = []proto.Session{{SessionID: "main", Assistant: true, Status: proto.StatusDone}}
	f.hello.Assistant.MemoryDreaming = true
	f.judge = `{"body":""}`
	st := &memStore{}
	h := &dreamTestHub{fakeHub: f}
	s := withSettings(t, f, New(Options{Hub: h, Store: func() Store { return st }, Logf: testLogf(t)}))
	setSettings(t, s, func(v *Settings) {
		v.Main, v.MainMachineID, v.MachineID = "main", "server", "server"
	})

	for _, tc := range []struct {
		name, answer, wantHistory string
		seq, wantSeq              int
	}{
		{"changed transcript", "replacement", "replacement", 2, 2},
		{"older node omits history", "", "original", 101, 101},
		{"same transcript", "original", "original", 2, 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := st.PutSetting(ctx, memoryDreamKey, memoryDreamProgress{
				Machine: "server", Session: "main", Seq: 100, History: "original", CheckedAt: time.Now().Add(-48 * time.Hour),
			}); err != nil {
				t.Fatal(err)
			}
			f.answerNode(DeviceSessionItems, func(_ string, _ json.RawMessage) (any, error) {
				return map[string]any{
					"history": tc.answer,
					"items":   []proto.Item{{Seq: tc.seq, Role: "user", Kind: proto.KindText, Text: "I prefer tea."}},
				}, nil
			})
			before := f.judged
			s.memoryDream(ctx)
			var progress memoryDreamProgress
			if _, err := st.Setting(ctx, memoryDreamKey, &progress); err != nil {
				t.Fatal(err)
			}
			if progress.Seq != tc.wantSeq || progress.History != tc.wantHistory {
				t.Fatalf("checkpoint = %+v", progress)
			}
			if called := f.judged > before; called != (tc.name != "same transcript") {
				t.Fatalf("model called = %v", called)
			}
		})
	}
}
