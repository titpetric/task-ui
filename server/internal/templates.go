package internal

import (
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"path/filepath"

	"github.com/apex/log"

	. "github.com/titpetric/task-ui/server/model"
)

const commonTemplate = "templates/_common.tpl"

func embeddedLoader(files *embed.FS) func(string, template.FuncMap) (*template.Template, error) {
	return func(filename string, funcs template.FuncMap) (*template.Template, error) {
		t, err := template.New(filepath.Base(filename)).Funcs(funcs).ParseFS(files, filename, commonTemplate)
		if err != nil {
			return nil, err
		}
		return t, nil
	}
}

func loadTemplateFromFilesystem(filename string, funcs template.FuncMap) (*template.Template, error) {
	t, err := template.New(filepath.Base(filename)).Funcs(funcs).ParseFiles(filename, commonTemplate)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func loadTemplate(name string, files *embed.FS, funcs template.FuncMap) (*template.Template, error) {
	loadTemplateFromEmbedFS := embeddedLoader(files)
	loaders := []func(string, template.FuncMap) (*template.Template, error){
		loadTemplateFromEmbedFS,
		loadTemplateFromFilesystem,
	}

	for _, loader := range loaders {
		t, err := loader(name, funcs)
		if err == nil {
			return t, nil
		}
		if !errors.Is(err, ErrNotFound) {
			log.Warnf("Loader error: %s", err)
		}
	}

	return nil, errors.New("no such template: " + name)
}

type TemplateRendererFunc func(w io.Writer, templateName string, data any) error

type TemplateFuncMap interface {
	FuncMap() template.FuncMap
}

func NewTemplateRenderer(files *embed.FS) TemplateRendererFunc {
	// default functions for the templates
	templateFuncs := template.FuncMap{
		"json": func(in interface{}) (string, error) {
			out, err := json.MarshalIndent(in, "", "  ")
			return string(out), err
		},
	}

	return func(w io.Writer, templateName string, data any) error {
		dataFuncs, ok := data.(TemplateFuncMap)
		if ok {
			fns := dataFuncs.FuncMap()
			for key, val := range fns {
				templateFuncs[key] = val
			}
		}

		t, err := loadTemplate(templateName, files, templateFuncs)
		if err != nil {
			return err
		}

		return t.Execute(w, data)
	}
}
