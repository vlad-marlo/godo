package model

import "github.com/google/uuid"

type (
	CreateInviteRequest struct {
		Group   uuid.UUID `json:"group" example:"00000000-0000-0000-0000-000000000000"`
		Limit   int       `json:"limit" example:"2"`
		Member  int       `json:"members-permission" example:"4"`
		Task    int       `json:"tasks-permission" example:"4"`
		Review  int       `json:"reviews-permission" example:"4"`
		Comment int       `json:"comments-permission" example:"4"`
	}
	CreateInviteResponse struct {
		Link  string `json:"invite-link" example:"http://localhost:8080/api/v1/groups/00000000-0000-0000-0000-000000000000/apply?invite=00000000-0000-0000-0000-000000000000"`
		Limit int    `json:"limit" example:"2"`
	}
	CreateInviteViaGroupRequest struct {
		Limit   int `json:"limit" example:"2"`
		Member  int `json:"members-permission" example:"4"`
		Task    int `json:"tasks-permission" example:"4"`
		Review  int `json:"reviews-permission" example:"4"`
		Comment int `json:"comments-permission" example:"4"`
	}
)
