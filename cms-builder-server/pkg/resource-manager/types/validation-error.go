package types

type ValidationError struct {
	Field string // The name of the field that failed validation
	Error string // The error message
}
