package cms_server

import (
	"html/template"
	"net/http"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	templateName := "base.html"
	templatePath := config.RootDir + "/templates/layouts/" + templateName

	t, err := template.New(templateName).ParseFiles(templatePath)
	if err != nil {
		panic(err)
	}
	t.Execute(w, nil)
}
