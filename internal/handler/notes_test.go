package handler

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/notes-app/internal/model"
	"github.com/notes-app/internal/service"
	"github.com/notes-app/internal/view"
)

type mockNoteService struct {
	indexFn      func(ctx context.Context, userID, page int, query string) (model.Pagination, error)
	createFn     func(ctx context.Context, userID int, form model.NoteForm) (*model.Note, error)
	getForUserFn func(ctx context.Context, noteID, userID int) (*model.Note, error)
	updateFn     func(ctx context.Context, noteID, userID int, form model.NoteForm) (*model.Note, error)
	deleteFn     func(ctx context.Context, noteID, userID int) error
	togglePinFn  func(ctx context.Context, noteID, userID int) (bool, error)
}

func (m *mockNoteService) Index(ctx context.Context, userID, page int, query string) (model.Pagination, error) {
	return m.indexFn(ctx, userID, page, query)
}
func (m *mockNoteService) Create(ctx context.Context, userID int, form model.NoteForm) (*model.Note, error) {
	return m.createFn(ctx, userID, form)
}
func (m *mockNoteService) GetForUser(ctx context.Context, noteID, userID int) (*model.Note, error) {
	return m.getForUserFn(ctx, noteID, userID)
}
func (m *mockNoteService) Update(ctx context.Context, noteID, userID int, form model.NoteForm) (*model.Note, error) {
	return m.updateFn(ctx, noteID, userID, form)
}
func (m *mockNoteService) Delete(ctx context.Context, noteID, userID int) error {
	return m.deleteFn(ctx, noteID, userID)
}
func (m *mockNoteService) TogglePin(ctx context.Context, noteID, userID int) (bool, error) {
	return m.togglePinFn(ctx, noteID, userID)
}

func setupTest() (*gin.Engine, *mockNoteService) {
	gin.SetMode(gin.TestMode)

	tmpl := template.Must(template.New("").Parse(
		"{{define \"base.html\"}}{{.content}}{{end}}{{define \"notes/index\"}}ok{{end}}{{define \"notes/form\"}}form{{end}}",
	))
	view.SetTemplates(tmpl)

	mock := &mockNoteService{
		indexFn: func(ctx context.Context, userID, page int, query string) (model.Pagination, error) {
			return model.Pagination{}, nil
		},
		createFn: func(ctx context.Context, userID int, form model.NoteForm) (*model.Note, error) {
			return &model.Note{ID: 1, UserID: userID, Title: form.Title, Content: form.Content, Color: form.Color}, nil
		},
		getForUserFn: func(ctx context.Context, noteID, userID int) (*model.Note, error) {
			return &model.Note{ID: noteID, UserID: userID, Title: "Test", Content: "Content", Color: "#6366f1", IsPinned: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
		updateFn: func(ctx context.Context, noteID, userID int, form model.NoteForm) (*model.Note, error) {
			return &model.Note{ID: noteID, UserID: userID, Title: form.Title, Content: form.Content, Color: form.Color}, nil
		},
		deleteFn: func(ctx context.Context, noteID, userID int) error {
			return nil
		},
		togglePinFn: func(ctx context.Context, noteID, userID int) (bool, error) {
			return true, nil
		},
	}
	h := NewNotesHandler(mock)
	r := gin.New()
	r.SetHTMLTemplate(tmpl)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	r.GET("/notes", h.Index)
	r.GET("/notes/create", h.Create)
	r.POST("/notes", h.Store)
	r.GET("/notes/:id/edit", h.Edit)
	r.POST("/notes/:id", h.Update)
	r.POST("/notes/:id/delete", h.Destroy)
	r.POST("/notes/:id/toggle-pin", h.TogglePin)
	return r, mock
}

func TestIndex_Success(t *testing.T) {
	r, mock := setupTest()
	mock.indexFn = func(ctx context.Context, userID, page int, query string) (model.Pagination, error) {
		return model.Pagination{
			Page: 1, PerPage: 20, Total: 1, TotalPages: 1,
			Items: []model.Note{{ID: 1, Title: "Note 1"}},
		}, nil
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notes", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestIndex_ServiceError(t *testing.T) {
	r, mock := setupTest()
	mock.indexFn = func(ctx context.Context, userID, page int, query string) (model.Pagination, error) {
		return model.Pagination{}, errors.New("db error")
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notes", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestCreate_Success(t *testing.T) {
	r, _ := setupTest()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notes/create", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestStore_Success(t *testing.T) {
	r, _ := setupTest()
	w := httptest.NewRecorder()
	body := strings.NewReader("title=Test+Note&content=Hello&color=%236366f1&_csrf=abc")
	req := httptest.NewRequest("POST", "/notes", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
}

func TestStore_ValidationError(t *testing.T) {
	r, mock := setupTest()
	mock.createFn = func(ctx context.Context, userID int, form model.NoteForm) (*model.Note, error) {
		return nil, service.ErrInvalidInput
	}
	w := httptest.NewRecorder()
	body := strings.NewReader("title=&_csrf=abc")
	req := httptest.NewRequest("POST", "/notes", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with form, got %d", w.Code)
	}
}

func TestEdit_Success(t *testing.T) {
	r, _ := setupTest()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notes/1/edit", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEdit_InvalidID(t *testing.T) {
	r, _ := setupTest()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notes/abc/edit", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestEdit_NotFound(t *testing.T) {
	r, mock := setupTest()
	mock.getForUserFn = func(ctx context.Context, noteID, userID int) (*model.Note, error) {
		return nil, service.ErrNotFound
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notes/999/edit", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestEdit_NotOwned(t *testing.T) {
	r, mock := setupTest()
	mock.getForUserFn = func(ctx context.Context, noteID, userID int) (*model.Note, error) {
		return nil, service.ErrNotOwned
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notes/1/edit", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdate_Success(t *testing.T) {
	r, _ := setupTest()
	w := httptest.NewRecorder()
	body := strings.NewReader("title=Updated&content=New&color=%23ef4444&_csrf=abc")
	req := httptest.NewRequest("POST", "/notes/1", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
}

func TestUpdate_InvalidID(t *testing.T) {
	r, _ := setupTest()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/notes/abc", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	r, mock := setupTest()
	mock.updateFn = func(ctx context.Context, noteID, userID int, form model.NoteForm) (*model.Note, error) {
		return nil, service.ErrNotFound
	}
	w := httptest.NewRecorder()
	body := strings.NewReader("title=Updated&_csrf=abc")
	req := httptest.NewRequest("POST", "/notes/999", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdate_NotOwned(t *testing.T) {
	r, mock := setupTest()
	mock.updateFn = func(ctx context.Context, noteID, userID int, form model.NoteForm) (*model.Note, error) {
		return nil, service.ErrNotOwned
	}
	w := httptest.NewRecorder()
	body := strings.NewReader("title=Updated&_csrf=abc")
	req := httptest.NewRequest("POST", "/notes/999", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDestroy_Success(t *testing.T) {
	r, _ := setupTest()
	w := httptest.NewRecorder()
	body := strings.NewReader("_csrf=abc")
	req := httptest.NewRequest("POST", "/notes/1/delete", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
}

func TestDestroy_InvalidID(t *testing.T) {
	r, _ := setupTest()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/notes/abc/delete", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDestroy_NotFound(t *testing.T) {
	r, mock := setupTest()
	mock.deleteFn = func(ctx context.Context, noteID, userID int) error {
		return service.ErrNotFound
	}
	w := httptest.NewRecorder()
	body := strings.NewReader("_csrf=abc")
	req := httptest.NewRequest("POST", "/notes/999/delete", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDestroy_NotOwned(t *testing.T) {
	r, mock := setupTest()
	mock.deleteFn = func(ctx context.Context, noteID, userID int) error {
		return service.ErrNotOwned
	}
	w := httptest.NewRecorder()
	body := strings.NewReader("_csrf=abc")
	req := httptest.NewRequest("POST", "/notes/999/delete", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTogglePin_Success(t *testing.T) {
	r, _ := setupTest()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/notes/1/toggle-pin", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if resp["ok"] != true {
		t.Error("expected ok=true")
	}
	if resp["is_pinned"] != true {
		t.Error("expected is_pinned=true")
	}
}

func TestTogglePin_InvalidID(t *testing.T) {
	r, _ := setupTest()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/notes/abc/toggle-pin", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTogglePin_NotFound(t *testing.T) {
	r, mock := setupTest()
	mock.togglePinFn = func(ctx context.Context, noteID, userID int) (bool, error) {
		return false, service.ErrNotFound
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/notes/999/toggle-pin", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTogglePin_NotOwned(t *testing.T) {
	r, mock := setupTest()
	mock.togglePinFn = func(ctx context.Context, noteID, userID int) (bool, error) {
		return false, service.ErrNotOwned
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/notes/999/toggle-pin", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
