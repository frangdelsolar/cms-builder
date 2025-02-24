package models

type File struct {
	*SystemData
	Name     string `json:"name"`
	Path     string `json:"path"` // relative path
	Url      string `json:"url"`  // absolute path
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}
