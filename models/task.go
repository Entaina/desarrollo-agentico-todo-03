package models

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TitleMaxLength es el número máximo de caracteres permitido en el título.
const TitleMaxLength = 200

// Prioridades admitidas para una tarea.
const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"

	// DefaultPriority es la prioridad de las tareas sin prioridad indicada.
	DefaultPriority = PriorityMedium
)

// Priorities lista las prioridades de menor a mayor, para renderizar opciones.
var Priorities = []string{PriorityLow, PriorityMedium, PriorityHigh}

// PriorityLabel devuelve la etiqueta en español de una prioridad.
func PriorityLabel(p string) string {
	switch p {
	case PriorityLow:
		return "Baja"
	case PriorityMedium:
		return "Media"
	case PriorityHigh:
		return "Alta"
	}
	return "Sin prioridad"
}

// Task representa una tarea de la lista.
type Task struct {
	ID        uuid.UUID `json:"id" db:"id" form:"-"`
	Title     string    `json:"title" db:"title" form:"title"`
	Completed bool      `json:"completed" db:"completed" form:"-"`
	Priority  string    `json:"priority" db:"priority" form:"priority"`
	CreatedAt time.Time `json:"created_at" db:"created_at" form:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" form:"-"`
}

// Tasks es una colección de Task.
type Tasks []Task

// Validate se ejecuta en cada llamada a los métodos "pop.Validate*".
func (t *Task) Validate(tx *pop.Connection) (*validate.Errors, error) {
	verrs := validate.NewErrors()

	title := strings.TrimSpace(t.Title)
	if title == "" {
		verrs.Add("title", "El título es obligatorio.")
	} else if len([]rune(title)) > TitleMaxLength {
		verrs.Add("title", fmt.Sprintf("El título no puede superar los %d caracteres.", TitleMaxLength))
	}

	if t.Priority != "" && !slices.Contains(Priorities, t.Priority) {
		verrs.Add("priority", "La prioridad no es válida.")
	}

	return verrs, nil
}

// BeforeSave asigna la prioridad por defecto cuando no se ha indicado ninguna.
func (t *Task) BeforeSave(tx *pop.Connection) error {
	if t.Priority == "" {
		t.Priority = DefaultPriority
	}
	return nil
}
