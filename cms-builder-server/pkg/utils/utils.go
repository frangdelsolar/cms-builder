package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gertd/go-pluralize"
)

// GetStructName returns the name of the given model, lowercased and without the package name.
//
// For example, if the input is a pointer to a struct named "User" in the "models" package,
// the method will return "user".
//
// Parameters:
// - a: The model to get the name for. It can be either a pointer to a struct or a struct directly.
//
// Returns:
// - string: The name of the model.
func GetStructName(a interface{}) string {
	modelName := fmt.Sprintf("%T", a)
	name := modelName[strings.LastIndex(modelName, ".")+1:]
	return name
}

// Pluralize returns the plural form of the given word.
//
// Parameters:
// - word: The word to pluralize.
//
// Returns:
// - string: The plural form of the word.
func Pluralize(word string) string {
	p := pluralize.NewClient()
	return p.Plural(word)
}

func SnakeCase(s string) string {
	re := regexp.MustCompile("(.)([A-Z][a-z]+)")
	s = re.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
}

func KebabCase(s string) string {
	re := regexp.MustCompile("(.)([A-Z][a-z]+)")
	s = re.ReplaceAllString(s, "${1}-${2}")
	return strings.ToLower(s)
}
