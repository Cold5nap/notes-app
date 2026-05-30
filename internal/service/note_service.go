package service

import (
	"context"
	"errors"
	"math"

	"github.com/notes-app/internal/model"
	"github.com/notes-app/internal/repository"
)

const PerPage = 20

var (
	ErrNotFound      = errors.New("note not found")
	ErrNotOwned      = errors.New("note does not belong to user")
	ErrInvalidInput  = errors.New("invalid input")
)

type NoteService struct {
	repo *repository.NoteRepo
}

func NewNoteService(repo *repository.NoteRepo) *NoteService {
	return &NoteService{repo: repo}
}

func (s *NoteService) Index(ctx context.Context, userID, page int, query string) (model.Pagination, error) {
	if page < 1 {
		page = 1
	}

	items, total, err := s.repo.List(ctx, userID, page, PerPage, query)
	if err != nil {
		return model.Pagination{}, err
	}

	return model.Pagination{
		Page:       page,
		PerPage:    PerPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(PerPage))),
		Items:      items,
		Query:      query,
	}, nil
}

func (s *NoteService) GetForUser(ctx context.Context, noteID, userID int) (*model.Note, error) {
	note, err := s.repo.GetByID(ctx, noteID)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, ErrNotFound
	}
	if note.UserID != userID {
		return nil, ErrNotOwned
	}
	return note, nil
}

func (s *NoteService) Create(ctx context.Context, userID int, form model.NoteForm) (*model.Note, error) {
	if errs := form.Validate(); len(errs) > 0 {
		return nil, ErrInvalidInput
	}

	color := form.Color
	if color == "" {
		color = "#6366f1"
	}

	note := &model.Note{
		UserID:   userID,
		Title:    form.Title,
		Content:  form.Content,
		Color:    color,
		IsPinned: false,
	}

	if err := s.repo.Create(ctx, note); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *NoteService) Update(ctx context.Context, noteID, userID int, form model.NoteForm) (*model.Note, error) {
	note, err := s.GetForUser(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	if errs := form.Validate(); len(errs) > 0 {
		return nil, ErrInvalidInput
	}

	color := form.Color
	if color == "" {
		color = "#6366f1"
	}

	note.Title = form.Title
	note.Content = form.Content
	note.Color = color

	if err := s.repo.Update(ctx, note); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *NoteService) Delete(ctx context.Context, noteID, userID int) error {
	note, err := s.GetForUser(ctx, noteID, userID)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, note.ID)
}

func (s *NoteService) TogglePin(ctx context.Context, noteID, userID int) (bool, error) {
	note, err := s.GetForUser(ctx, noteID, userID)
	if err != nil {
		return false, err
	}
	return s.repo.TogglePin(ctx, note.ID)
}
