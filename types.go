package main

import "encoding/json"

// === Tool Definitions ===

type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

type ToolContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// === Extension Messages ===

type ExtensionMessage struct {
	Type    string          `json:"type"`
	Tools   []ToolDefinition `json:"tools,omitempty"`
	Names   []string        `json:"names,omitempty"`
	CallID  string          `json:"callId,omitempty"`
	Result  *ToolResult     `json:"result,omitempty"`
}

type CallToolMessage struct {
	Type   string                 `json:"type"`
	CallID string                 `json:"callId"`
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

// === JSON-RPC 2.0 ===

type JsonRpcRequest struct {
	Jsonrpc string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type JsonRpcResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JsonRpcError `json:"error,omitempty"`
}

type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Raw JSON unmarshaler helper
type rawJSON = json.RawMessage
