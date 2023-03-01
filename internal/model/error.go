package model

// Error is abstract error response type.
type Error struct {
	Err   string `json:"error" example:"short summary info about error"`
	Field string `json:"field" example:"additional info about error"`
}
