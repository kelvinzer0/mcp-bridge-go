package room

import (
	"testing"

	"github.com/kelvinzer0/mcp-bridge-go/internal/protocol"
)

func TestRoomToolsRegistration(t *testing.T) {
	r := New("test-room")

	if r.IsConnected() {
		t.Error("expected room to not be connected initially")
	}

	tools := []protocol.ToolDefinition{
		{
			Name:        "get_weather",
			Description: "Get weather information",
			InputSchema: map[string]interface{}{"type": "object"},
		},
		{
			Name:        "execute_bash",
			Description: "Execute a bash command",
		},
	}

	r.RegisterTools(tools)

	if !r.HasTool("get_weather") {
		t.Error("expected room to have tool 'get_weather'")
	}
	if !r.HasTool("execute_bash") {
		t.Error("expected room to have tool 'execute_bash'")
	}

	list := r.GetTools()
	if len(list) != 2 {
		t.Errorf("expected 2 tools, got %d", len(list))
	}

	// Test unregister
	r.UnregisterTools([]string{"execute_bash"})
	if r.HasTool("execute_bash") {
		t.Error("expected 'execute_bash' to be unregistered")
	}
	if len(r.GetTools()) != 1 {
		t.Errorf("expected 1 tool, got %d", len(r.GetTools()))
	}
}

func TestHubRegistry(t *testing.T) {
	hub := NewHub()

	r1 := hub.GetOrCreate("room-a")
	if r1 == nil {
		t.Fatal("expected room-a to be created")
	}

	r2 := hub.Get("room-a")
	if r1 != r2 {
		t.Error("expected same room instance from Get")
	}

	if hub.Get("non-existent") != nil {
		t.Error("expected nil for non-existent room")
	}
}
