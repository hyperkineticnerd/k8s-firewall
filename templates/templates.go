package templates

import (
	"bytes"
	"text/template"

	"github.com/hyperkineticnerd/k8s-firewall/config"
	"github.com/hyperkineticnerd/k8s-firewall/provider/juniper"
)

type TemplateEngine struct {
	Templates *template.Template
}

func TemplateSetup(cfg *config.Config) *TemplateEngine {
	tmpl, err := template.New(cfg.TemplateName).ParseFiles(cfg.TemplatePath + cfg.TemplateName)
	if err != nil {
		panic(err)
	}
	return &TemplateEngine{
		Templates: tmpl,
	}
}

func (t *TemplateEngine) TemplateRender(data *juniper.PortForward) (string, error) {
	var output bytes.Buffer
	e := t.Templates.Execute(&output, &data)
	if e != nil {
		panic(e)
	}
	return output.String(), nil
}
