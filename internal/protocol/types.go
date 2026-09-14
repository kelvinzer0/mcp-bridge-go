package protocol

// ToolDefinition defines the structure of a tool exposed by the extension.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// ToolContent represents a content element returned by a tool execution.
type ToolContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// ToolResult represents the output of a tool call.
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ExtensionMessage represents an inbound WebSocket payload from the browser extension.
type ExtensionMessage struct {
	Type   string           `json:"type"`
	Tools  []ToolDefinition `json:"tools,omitempty"`
	Names  []string         `json:"names,omitempty"`
	CallID string           `json:"callId,omitempty"`
	Result *ToolResult      `json:"result,omitempty"`
}

// CallToolMessage represents an outbound tool call sent from the bridge to the extension.
type CallToolMessage struct {
	Type   string                 `json:"type"`
	CallID string                 `json:"callId"`
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

// JsonRpcRequest represents a JSON-RPC 2.0 request payload.
type JsonRpcRequest struct {
	Jsonrpc string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// JsonRpcResponse represents a JSON-RPC 2.0 response payload.
type JsonRpcResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JsonRpcError `json:"error,omitempty"`
}

// JsonRpcError represents a JSON-RPC 2.0 error detail.
type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
