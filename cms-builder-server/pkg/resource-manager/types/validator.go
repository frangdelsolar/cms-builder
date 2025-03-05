package types

type Validator func(fieldName string, entity EntityData, output *ValidationError) *ValidationError
