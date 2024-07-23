package cms_server

import (
	"html/template"
	"path/filepath"
)

var templates map[string]*template.Template

type TemplateConfig struct {
	TemplateLayoutPath  string
	TemplateIncludePath string
}

var templateConfig TemplateConfig
var mainTmpl = `{{define "main" }} {{ template "base" . }} {{ end }}`

func LoadTemplateConfiguration() {
	templateConfig.TemplateIncludePath = filepath.Join(config.RootDir, "templates")
	templateConfig.TemplateLayoutPath = filepath.Join(templateConfig.TemplateIncludePath, "layouts")
}

func LoadTemplates() {
	log.Debug().Msg("Loading templates")

	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	layoutFiles, err := filepath.Glob(templateConfig.TemplateLayoutPath + "*.html")
	if err != nil {
		log.Error().Err(err).Msg("Error loading layout templates")
	}

	includeFiles, err := filepath.Glob(templateConfig.TemplateIncludePath + "*.html")
	if err != nil {
		log.Error().Err(err).Msg("Error loading include templates")
	}

	mainTemplate := template.New("main")

	mainTemplate, err = mainTemplate.Parse(mainTmpl)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing main template")
	}

	for _, file := range includeFiles {
		fileName := filepath.Base(file)
		files := append(layoutFiles, file)
		templates[fileName], err = mainTemplate.Clone()
		if err != nil {
			log.Error().Err(err).Msg("Error cloning template")
		}
		templates[fileName] = template.Must(templates[fileName].ParseFiles(files...))

		log.Debug().Msgf("Loaded template %s", fileName)
	}

}
