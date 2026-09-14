package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kelvinzer0/mcp-bridge-go/internal/protocol"
	"github.com/kelvinzer0/mcp-bridge-go/internal/room"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  65536,
	WriteBufferSize: 65536,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub *room.Hub
}

func NewHandler(hub *room.Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) getBaseURLs(r *http.Request) (httpBase, wsBase string) {
	proto := "http"
	wsProto := "ws"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		proto = "https"
		wsProto = "wss"
	}

	host := r.Host
	if fHost := r.Header.Get("X-Forwarded-Host"); fHost != "" {
		host = fHost
	}

	return fmt.Sprintf("%s://%s", proto, host), fmt.Sprintf("%s://%s", wsProto, host)
}

func generateRandomID(length int) string {
	bytes := make([]byte, length/2+1)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())[:length]
	}
	return hex.EncodeToString(bytes)[:length]
}

func (h *Handler) HandleNew(w http.ResponseWriter, r *http.Request) {
	roomID := generateRandomID(8)
	_ = h.hub.GetOrCreate(roomID)

	httpBase, wsBase := h.getBaseURLs(r)

	resp := map[string]string{
		"room":          roomID,
		"extension_url": fmt.Sprintf("%s/ws/extension?room=%s", wsBase, roomID),
		"mcp_url":       fmt.Sprintf("%s/mcp?room=%s", httpBase, roomID),
		"health_url":    fmt.Sprintf("%s/health?room=%s", httpBase, roomID),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		roomID = "default"
	}

	rInstance := h.hub.Get(roomID)
	isConnected := false
	toolsRegistered := 0
	toolNames := []string{}

	if rInstance != nil {
		isConnected = rInstance.IsConnected()
		tools := rInstance.GetTools()
		toolsRegistered = len(tools)
		for _, t := range tools {
			toolNames = append(toolNames, t.Name)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":             "ok",
		"extensionConnected": isConnected,
		"toolsRegistered":    toolsRegistered,
		"tools":              toolNames,
	})
}

func (h *Handler) HandleWSExtension(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		roomID = "default"
	}

	rInstance := h.hub.GetOrCreate(roomID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade failed: %v", err)
		return
	}

	rInstance.SetWS(conn)
	defer func() {
		rInstance.ClearWS(conn)
		_ = conn.Close()
	}()

	log.Printf("[WS] Extension connected to room '%s'", roomID)

	done := make(chan struct{})
	defer close(done)

	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := rInstance.SendWSJSON(map[string]string{"type": "ping"}); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[WS] Unexpected close: %v", err)
			}
			break
		}

		var msg protocol.ExtensionMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "registerTools":
			if len(msg.Tools) > 0 {
				rInstance.RegisterTools(msg.Tools)
				log.Printf("[Room %s] Registered %d tools", roomID, len(msg.Tools))
			}
		case "unregisterTools":
			if len(msg.Names) > 0 {
				rInstance.UnregisterTools(msg.Names)
				log.Printf("[Room %s] Unregistered %d tools", roomID, len(msg.Names))
			}
		case "toolResult":
			if msg.CallID != "" && msg.Result != nil {
				rInstance.HandleToolResult(msg.CallID, msg.Result)
			}
		case "pong":
			// extension heartbeat
		}
	}

	log.Printf("[WS] Extension disconnected from room '%s'", roomID)
}

func (h *Handler) HandleMCP(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		roomID = "default"
	}
	rInstance := h.hub.GetOrCreate(roomID)

	switch r.Method {
	case http.MethodGet:
		h.handleMCPSSE(w, r, rInstance, roomID)
	case http.MethodPost:
		h.handleMCPRPC(w, r, rInstance)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleMCPSSE(w http.ResponseWriter, r *http.Request, rInstance *room.Room, roomID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	endpointData := fmt.Sprintf("/mcp?room=%s", roomID)
	_, _ = fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpointData)
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			_, err := fmt.Fprintf(w, ":\n\n")
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *Handler) handleMCPRPC(w http.ResponseWriter, r *http.Request, rInstance *room.Room) {
	w.Header().Set("Content-Type", "application/json")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		_ = json.NewEncoder(w).Encode(protocol.JsonRpcResponse{
			Jsonrpc: "2.0",
			ID:      nil,
			Error:   &protocol.JsonRpcError{Code: -32700, Message: "Parse error"},
		})
		return
	}

	var req protocol.JsonRpcRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		_ = json.NewEncoder(w).Encode(protocol.JsonRpcResponse{
			Jsonrpc: "2.0",
			ID:      nil,
			Error:   &protocol.JsonRpcError{Code: -32700, Message: "Parse error"},
		})
		return
	}

	if req.ID == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	resp := protocol.JsonRpcResponse{
		Jsonrpc: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]string{
				"name":    "mcp-bridge",
				"version": "1.0.0",
			},
		}

	case "tools/list":
		tools := rInstance.GetTools()
		toolList := make([]map[string]interface{}, 0, len(tools))
		for _, t := range tools {
			schema := t.InputSchema
			if schema == nil {
				schema = map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				}
			}
			toolList = append(toolList, map[string]interface{}{
				"name":        t.Name,
				"description": t.Description,
				"inputSchema": schema,
			})
		}
		resp.Result = map[string]interface{}{
			"tools": toolList,
		}

	case "tools/call":
		toolName, _ := req.Params["name"].(string)
		args, _ := req.Params["arguments"].(map[string]interface{})
		if args == nil {
			args = make(map[string]interface{})
		}

		if toolName == "" || !rInstance.HasTool(toolName) {
			resp.Error = &protocol.JsonRpcError{
				Code:    -32602,
				Message: fmt.Sprintf("Unknown tool: %s", toolName),
			}
			break
		}

		if !rInstance.IsConnected() {
			resp.Result = protocol.ToolResult{
				Content: []protocol.ToolContent{
					{Type: "text", Text: "Extension not connected"},
				},
				IsError: true,
			}
			break
		}

		res, err := rInstance.CallTool(r.Context(), toolName, args, 60*time.Second)
		if err != nil {
			resp.Result = protocol.ToolResult{
				Content: []protocol.ToolContent{
					{Type: "text", Text: fmt.Sprintf("Error: %s", err.Error())},
				},
				IsError: true,
			}
		} else {
			resp.Result = res
		}

	case "ping":
		resp.Result = map[string]interface{}{}

	default:
		resp.Error = &protocol.JsonRpcError{
			Code:    -32601,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
	}

	_ = json.NewEncoder(w).Encode(resp)
}
