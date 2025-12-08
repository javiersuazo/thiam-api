package response

type Error struct {
	Code    string            `json:"code" example:"VALIDATION_ERROR"`
	Message string            `json:"message" example:"Invalid request parameters"`
	Details map[string]string `json:"details,omitempty"`
}
