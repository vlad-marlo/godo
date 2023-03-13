package model

import "github.com/google/uuid"

type (
	// CreateInviteRequest represents data that must be passed by user to create invite.
	CreateInviteRequest struct {
		Group   uuid.UUID `json:"group" example:"00000000-0000-0000-0000-000000000000"`
		Limit   int       `json:"limit" example:"2"`
		Member  int       `json:"members-permission" example:"4"`
		Task    int       `json:"tasks-permission" example:"4"`
		Review  int       `json:"reviews-permission" example:"4"`
		Comment int       `json:"comments-permission" example:"4"`
	}
	// CreateInviteRequest is response returned to user.
	CreateInviteResponse struct {
		// Link is invite link to group.
		Link string `json:"invite-link" example:"http://localhost:8080/api/v1/groups/00000000-0000-0000-0000-000000000000/apply?invite=00000000-0000-0000-0000-000000000000"`
		// Limit is count of avaliable usages of invite link.
		Limit int `json:"limit" example:"2"`
	}
	// CreateInviteViaGroupRequest is response object to create invite.
	CreateInviteViaGroupRequest struct {
		// Limit is count of avaliable usages of invite link
		Limit int `json:"limit" example:"2"`
		// Member is role of user.
		//
		// There are permissions:
		// 0 - user can affect(read/update) only related objects;
		// 1 - user can read all object.
		// 2 - user can create new objects (invite users);
		// 3 - user can change/delete users who was invited by whose;
		// 4 - user can affect all users.
		Member int `json:"members-permission" example:"4"`
		// Task is role of user.
		//
		// Permissions:
		// 0 - read related;
		// 1 - read all;
		// 2 create;
		// 3 - update/delete related;
		// 4 - delete all;
		Task int `json:"tasks-permission" example:"4"`
		// Review is role of user.
		//
		// Permissions:
		// 0 - read related;
		// 1 - read all;
		// 2 - create;
		// 3 - update/delete related;
		// 4 - update/delete any;
		Review int `json:"reviews-permission" example:"4"`
		// Comment is role of user.
		//
		// Permissions:
		// 0 - read related;
		// 1 - read all;
		// 2 - create;
		// 3 - update/delete related;
		// 4 - update/delete any;
		Comment int `json:"comments-permission" example:"4"`
	}
)
