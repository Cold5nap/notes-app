package model

import (
	"regexp"
	"strings"
	"time"
)

type Note struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Color     string    `json:"color"`
	IsPinned  bool      `json:"is_pinned"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type NoteForm struct {
	Title   string `form:"title"`
	Content string `form:"content"`
	Color   string `form:"color"`
}

type ValidationErrors map[string]string

func (f *NoteForm) Validate() ValidationErrors {
	errs := make(ValidationErrors)

	f.Title = strings.TrimSpace(f.Title)

	if f.Title == "" {
		errs["title"] = "Заголовок обязателен"
	} else if len(f.Title) > 255 {
		errs["title"] = "Заголовок не может быть длиннее 255 символов"
	}

	if f.Color != "" && !hexColor.MatchString(f.Color) {
		errs["color"] = "Цвет должен быть валидным hex-кодом"
	}

	return errs
}
