package main

import (
	"gorm.io/gorm"
)

type Primary struct {
	gorm.Model
	StringField string  `json:"string_field"`
	NumberField int     `json:"number_field"`
	BoolField   bool    `json:"bool_field"`
	FloatField  float64 `json:"float_field"`
}
