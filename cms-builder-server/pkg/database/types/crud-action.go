package types

type CRUDAction string

const (
	CreateCRUDAction CRUDAction = "created"
	UpdateCRUDAction CRUDAction = "updated"
	DeleteCRUDAction CRUDAction = "deleted"
)
