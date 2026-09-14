package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kelvinzer0/mcp-bridge-go/internal/protocol"
	"github.com/kelvinzer0/mcp-bridge-go/internal/room"
)

func TestHandleNew(t *testing.T) {
	hub := room.NewHub()
	handler := NewHandler(hub)

	req := httptest.NewRequest(http.MethodGet, "/new", nil)
	w := httptest.NewRecorder()

	handler.HandleNew(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["room"] == "" {
		t.Error("expected room id to be generated")
	}
	if resp["mcp_url"] == "" || resp["extension_url"] == "" {
		t.Error("expected connection URLs to be provided")
	}
}

func TestHandleHealth(t *testing.T) {
	hub := room.NewHub()
	handler := NewHandler(hub)

	r := hub.GetOrCreate("test-room")
	r.RegisterTools([]protocol.ToolDefinition{{Name: "calculator"}})

	req := httptest.NewRequest(http.MethodGet, "/health?room=test-room", nil)
	w := httptest.NewRecorder()

	handler.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}
	if resp["toolsRegistered"] != float64(1) {
		t.Errorf("expected 1 registered tool, got %v", resp["toolsRegistered"])
	}
}

func TestHandleMCPInitialize(t *testing.T) {
	hub := room.NewHub()
	handler := NewHandler(hub)

	payload := protocol.JsonRpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/mcp?room=test-room", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleMCP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp protocol.JsonRpcResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON-RPC: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected RPC error: %v", resp.Error.Message)
	}

	resMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object, got %T", resp.Result)
	}

	if resMap["protocolVersion"] != "2025-03-26" {
		t.Errorf("expected protocolVersion 2025-03-26, got %v", resMap["protocolVersion"])
	}
}

func TestHandleMCPUnknownMethod(t *testing.T) {
	hub := room.NewHub()
	handler := NewHandler(hub)

	payload := protocol.JsonRpcRequest{
		Jsonrpc: "2.0",
		ID:      2,
		Method:  "unknown_service_endpoint",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/mcp?room=test-room", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleMCP(w, req)

	var resp protocol.JsonRpcResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp.Error == nil || resp.Error.Code != -32601 {
		t.Errorf("expected JSON-RPC error code -32601, got %v", resp.Error)
	}
}
