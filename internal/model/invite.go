package model

type (
	// InviteUserRequest ...
	InviteUserRequest struct {
		// User is email or id of user.
		User string `json:"user"`
		// Limit is limit of usages of invite link.
		Limit int `json:"limit"`
	}
	InviteUserResponse struct {
	}
)
