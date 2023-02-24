package model

import "github.com/google/uuid"

type (
	Role struct {
		ID            int
		CreateTask    bool
		ReadTask      bool
		UpdateTask    bool
		DeleteTask    bool
		CreateIssue   bool
		ReadIssue     bool
		UpdateIssue   bool
		ReviewTask    bool
		ReadMembers   bool
		InviteMembers bool
		DeleteMembers bool
	}
	UserInGroup struct {
		User  uuid.UUID
		Group uuid.UUID
		Role  int
	}
)
