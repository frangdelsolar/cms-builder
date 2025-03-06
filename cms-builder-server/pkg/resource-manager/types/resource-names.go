package types

type ResourceNames struct {
	Singular      string `json:"singularName"`
	Plural        string `json:"pluralName"`
	SnakeSingular string `json:"snakeName"`
	SnakePlural   string `json:"snakePluralName"`
	KebabSingular string `json:"kebabName"`
	KebabPlural   string `json:"kebabPluralName"`
}
