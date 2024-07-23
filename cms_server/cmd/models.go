package main

type Primitive struct {
	StringField string  `json:"stringField"`
	NumberField int     `json:"numberField"`
	BoolField   bool    `json:"boolField"`
	FloatField  float64 `json:"floatField"`
}
