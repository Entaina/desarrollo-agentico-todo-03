package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TitleMaxLength es el número máximo de caracteres permitido en el título.
const TitleMaxLength = 200

// Niveles de prioridad. Media es el valor cero para que sea el valor por defecto.
const (
	PriorityLow    = -1
	PriorityMedium = 0
	PriorityHigh   = 1
)

// Task representa una tarea de la lista.
type Task struct {
	ID        uuid.UUID `json:"id" db:"id" form:"-"`
	Title     string    `json:"title" db:"title" form:"title"`
	Completed bool      `json:"completed" db:"completed" form:"-"`
	Priority  int       `json:"priority" db:"priority" form:"-"`
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

	if t.Priority < PriorityLow || t.Priority > PriorityHigh {
		verrs.Add("priority", "La prioridad no es válida.")
	}

	return verrs, nil
}

// PriorityKey devuelve el identificador de la prioridad para clases CSS y plantillas.
func (t Task) PriorityKey() string {
	switch t.Priority {
	case PriorityHigh:
		return "high"
	case PriorityLow:
		return "low"
	default:
		return "medium"
	}
}

// PriorityLabel devuelve el nombre legible de la prioridad.
func (t Task) PriorityLabel() string {
	switch t.Priority {
	case PriorityHigh:
		return "alta"
	case PriorityLow:
		return "baja"
	default:
		return "media"
	}
}
