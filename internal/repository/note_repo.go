package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/notes-app/internal/model"
)

type NoteRepo struct {
	db *sql.DB
}

func NewNoteRepo(db *sql.DB) *NoteRepo {
	return &NoteRepo{db: db}
}

func (r *NoteRepo) List(ctx context.Context, userID int, page, perPage int, query string) ([]model.Note, int, error) {
	var total int
	countSQL := "SELECT COUNT(*) FROM notes WHERE user_id = $1"
	args := []any{userID}

	if query != "" {
		countSQL += " AND title ILIKE '%' || $2 || '%'"
		args = append(args, query)
	}

	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notes: %w", err)
	}

	dataSQL := `SELECT id, user_id, title, content, color, is_pinned, created_at, updated_at
		FROM notes WHERE user_id = $1`
	dataArgs := []any{userID}

	if query != "" {
		dataSQL += " AND title ILIKE '%' || $2 || '%'"
		dataArgs = append(dataArgs, query)
	}

	dataSQL += " ORDER BY is_pinned DESC, updated_at DESC LIMIT $" + fmt.Sprintf("%d", len(dataArgs)+1)
	dataSQL += " OFFSET $" + fmt.Sprintf("%d", len(dataArgs)+2)
	dataArgs = append(dataArgs, perPage, (page-1)*perPage)

	rows, err := r.db.QueryContext(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	var notes []model.Note
	for rows.Next() {
		var n model.Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.Color, &n.IsPinned, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, n)
	}

	return notes, total, rows.Err()
}

func (r *NoteRepo) GetByID(ctx context.Context, id int) (*model.Note, error) {
	n := &model.Note{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, title, content, color, is_pinned, created_at, updated_at FROM notes WHERE id = $1", id,
	).Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.Color, &n.IsPinned, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get note: %w", err)
	}
	return n, nil
}

func (r *NoteRepo) Create(ctx context.Context, n *model.Note) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO notes (user_id, title, content, color)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		n.UserID, n.Title, n.Content, n.Color,
	).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create note: %w", err)
	}
	return nil
}

func (r *NoteRepo) Update(ctx context.Context, n *model.Note) error {
	n.UpdatedAt = n.CreatedAt // will be set by RETURNING
	err := r.db.QueryRowContext(ctx,
		`UPDATE notes SET title = $1, content = $2, color = $3, updated_at = NOW()
		 WHERE id = $4
		 RETURNING updated_at`,
		n.Title, n.Content, n.Color, n.ID,
	).Scan(&n.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}
	return nil
}

func (r *NoteRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM notes WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	return nil
}

func (r *NoteRepo) TogglePin(ctx context.Context, id int) (bool, error) {
	var pinned bool
	err := r.db.QueryRowContext(ctx,
		`UPDATE notes SET is_pinned = NOT is_pinned, updated_at = NOW()
		 WHERE id = $1
		 RETURNING is_pinned`, id,
	).Scan(&pinned)
	if err != nil {
		return false, fmt.Errorf("toggle pin: %w", err)
	}
	return pinned, nil
}

// BuildSearchCondition returns a WHERE clause and args for optional title search.
// Used by List, but kept separate for clarity.
func buildSearchCondition(query string, startIdx int) (string, []any) {
	q := strings.TrimSpace(query)
	if q == "" {
		return "", nil
	}
	return fmt.Sprintf(" AND title ILIKE '%%' || $%d || '%%'", startIdx), []any{q}
}
