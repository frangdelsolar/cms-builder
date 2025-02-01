package main

import (
	"encoding/json"
	"fmt"

	"github.com/frangdelsolar/cms/builder"
	"github.com/invopop/jsonschema"
)

func main() {
	s := jsonschema.Reflect(builder.User{})

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(string(data))
}
