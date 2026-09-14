package room

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kelvinzer0/mcp-bridge-go/internal/protocol"
)

type Room struct {
	ID           string
	mu           sync.RWMutex
	tools        map[string]protocol.ToolDefinition
	ws           *websocket.Conn
	wsMu         sync.Mutex
	pendingCalls map[string]chan *protocol.ToolResult
}

func New(id string) *Room {
	return &Room{
		ID:           id,
		tools:        make(map[string]protocol.ToolDefinition),
		pendingCalls: make(map[string]chan *protocol.ToolResult),
	}
}

func (r *Room) SetWS(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ws != nil {
		_ = r.ws.Close()
	}
	r.ws = conn
}

func (r *Room) ClearWS(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ws == conn {
		r.ws = nil
		r.tools = make(map[string]protocol.ToolDefinition)

		for id, ch := range r.pendingCalls {
			close(ch)
			delete(r.pendingCalls, id)
		}
	}
}

func (r *Room) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ws != nil
}

func (r *Room) RegisterTools(tools []protocol.ToolDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range tools {
		r.tools[t.Name] = t
	}
}

func (r *Room) UnregisterTools(names []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, name := range names {
		delete(r.tools, name)
	}
}

func (r *Room) GetTools() []protocol.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]protocol.ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

func (r *Room) HasTool(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.tools[name]
	return exists
}

func (r *Room) SendWSJSON(v interface{}) error {
	r.wsMu.Lock()
	defer r.wsMu.Unlock()

	r.mu.RLock()
	conn := r.ws
	r.mu.RUnlock()

	if conn == nil {
		return errors.New("extension not connected")
	}

	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteJSON(v)
}

func (r *Room) CallTool(ctx context.Context, name string, args map[string]interface{}, timeout time.Duration) (*protocol.ToolResult, error) {
	if !r.IsConnected() {
		return nil, errors.New("extension not connected")
	}

	callID := uuid.NewString()
	resChan := make(chan *protocol.ToolResult, 1)

	r.mu.Lock()
	r.pendingCalls[callID] = resChan
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		delete(r.pendingCalls, callID)
		r.mu.Unlock()
	}()

	msg := protocol.CallToolMessage{
		Type:   "callTool",
		CallID: callID,
		Name:   name,
		Params: args,
	}

	if err := r.SendWSJSON(msg); err != nil {
		return nil, fmt.Errorf("failed to dispatch tool call: %w", err)
	}

	select {
	case res, ok := <-resChan:
		if !ok || res == nil {
			return nil, errors.New("extension disconnected")
		}
		return res, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout executing tool: %s", name)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *Room) HandleToolResult(callID string, res *protocol.ToolResult) {
	r.mu.RLock()
	ch, exists := r.pendingCalls[callID]
	r.mu.RUnlock()

	if exists && ch != nil {
		select {
		case ch <- res:
		default:
		}
	}
}
