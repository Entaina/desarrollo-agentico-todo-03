package actions

import (
	"slices"

	"todo/models"
	"todo/public"
	"todo/templates"

	"github.com/gobuffalo/buffalo/render"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.plush.html",

		// fs.FS containing templates
		TemplatesFS: templates.FS(),

		// fs.FS containing assets
		AssetsFS: public.FS(),

		// Add template helpers here:
		Helpers: render.Helpers{
			"priorities":     func() []string { return models.Priorities },
			"priority_label": models.PriorityLabel,
			"priority_class": priorityClass,

			// for non-bootstrap form helpers uncomment the lines
			// below and import "github.com/gobuffalo/helpers/forms"
			// forms.FormKey:     forms.Form,
			// forms.FormForKey:  forms.FormFor,
		},
	})
}

// priorityClass devuelve un sufijo de clase CSS seguro para una prioridad.
func priorityClass(p string) string {
	if slices.Contains(models.Priorities, p) {
		return p
	}
	return "none"
}
