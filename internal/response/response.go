package response

import "encoding/json"

// Envelope is a JSON response container for both success and error outcomes.
type Envelope struct {
	OK      bool   `json:"ok"`
	Data    any    `json:"data,omitempty"`
	ErrCode string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Success creates a success envelope with the given data.
func Success(data any) Envelope {
	return Envelope{OK: true, Data: data}
}

// Error creates an error envelope with the given error code and message.
func Error(code, message string) Envelope {
	return Envelope{OK: false, ErrCode: code, Message: message}
}

// JSON marshals the envelope to a JSON string. If marshaling fails, returns
// a safe error response.
func (e Envelope) JSON() string {
	b, err := json.Marshal(e)
	if err != nil {
		return `{"ok":false,"error":"internal_error","message":"failed to encode response"}`
	}
	return string(b)
}
