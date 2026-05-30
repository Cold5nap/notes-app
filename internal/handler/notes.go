package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/notes-app/internal/middleware"
	"github.com/notes-app/internal/model"
	"github.com/notes-app/internal/service"
	"github.com/notes-app/internal/view"
)

type NotesHandler struct {
	svc *service.NoteService
}

func NewNotesHandler(svc *service.NoteService) *NotesHandler {
	return &NotesHandler{svc: svc}
}

func (h *NotesHandler) Index(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	query := c.Query("q")

	pagination, err := h.svc.Index(c.Request.Context(), userID, page, query)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	view.Render(c, "notes/index", pagination)
}

func (h *NotesHandler) Create(c *gin.Context) {
	view.Render(c, "notes/form", gin.H{
		"note":   nil,
		"errors": nil,
	})
}

func (h *NotesHandler) Store(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var form model.NoteForm
	if err := c.ShouldBind(&form); err != nil {
		view.Render(c, "notes/form", gin.H{
			"note":   form,
			"errors": model.ValidationErrors{"_form": "Неверные данные формы"},
		})
		return
	}

	note, err := h.svc.Create(c.Request.Context(), userID, form)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			view.Render(c, "notes/form", gin.H{
				"note":   form,
				"errors": form.Validate(),
			})
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	view.SetFlash(c, "Заметка создана", "success")
	c.Redirect(http.StatusFound, "/notes")
	_ = note
}

func (h *NotesHandler) Edit(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	note, err := h.svc.GetForUser(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) || errors.Is(err, service.ErrNotOwned) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	view.Render(c, "notes/form", gin.H{
		"note":   note,
		"errors": nil,
	})
}

func (h *NotesHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var form model.NoteForm
	if err := c.ShouldBind(&form); err != nil {
		view.Render(c, "notes/form", gin.H{
			"note":   form,
			"errors": model.ValidationErrors{"_form": "Неверные данные формы"},
		})
		return
	}

	note, err := h.svc.Update(c.Request.Context(), id, userID, form)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) || errors.Is(err, service.ErrNotOwned) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			view.Render(c, "notes/form", gin.H{
				"note":   form,
				"errors": form.Validate(),
			})
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	view.SetFlash(c, "Заметка обновлена", "success")
	c.Redirect(http.StatusFound, "/notes")
	_ = note
}

func (h *NotesHandler) Destroy(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, service.ErrNotFound) || errors.Is(err, service.ErrNotOwned) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	view.SetFlash(c, "Заметка удалена", "success")
	c.Redirect(http.StatusFound, "/notes")
}

func (h *NotesHandler) TogglePin(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	pinned, err := h.svc.TogglePin(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) || errors.Is(err, service.ErrNotOwned) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"is_pinned": pinned,
	})
}
