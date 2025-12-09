package response

// Error represents a structured error response.
// Note: This is a new error format. For backwards compatibility with clients
// expecting the legacy {"error": "message"} format, use LegacyError instead.
type Error struct {
	Code    string            `json:"code" example:"VALIDATION_ERROR"`
	Message string            `json:"message" example:"Invalid request parameters"`
	Details map[string]string `json:"details,omitempty"`
}

// LegacyError represents the legacy error response format for backwards compatibility.
type LegacyError struct {
	Error string `json:"error" example:"An error occurred"`
}
