package model

const (
	PermReadRelated = iota
	PermReadAll
	PermCreate
	PermChangeRelated
	PermChangeAll
)

type (
	// Role is helper struct to store role.
	Role struct {
		ID       int32
		Members  int
		Tasks    int
		Reviews  int
		Comments int
	}
)
