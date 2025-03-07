package server

type QueryParams struct {
	Limit int               `json:"limit"`
	Page  int               `json:"page"`
	Order string            `json:"order"`
	Query map[string]string `json:"query"`
}
