package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/notes-app/internal/config"
	"github.com/notes-app/internal/handler"
	"github.com/notes-app/internal/middleware"
	"github.com/notes-app/internal/repository"
	"github.com/notes-app/internal/service"
	"github.com/notes-app/internal/view"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	noteRepo := repository.NewNoteRepo(db)
	noteSvc := service.NewNoteService(noteRepo)
	noteH := handler.NewNotesHandler(noteSvc)

	middleware.SetCookieSecure(cfg.CookieSecure)

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Static("/static", "./static")
	tmpl := template.New("").Funcs(template.FuncMap{
		"sub": func(a, b int) int { return a - b },
	})
	tmpl = template.Must(tmpl.ParseGlob("templates/layout/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/notes/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))
	r.SetHTMLTemplate(tmpl)
	view.SetTemplates(tmpl)

	r.GET("/", middleware.AuthRequired(), func(c *gin.Context) {
		c.Redirect(302, "/notes")
	})

	r.GET("/login", func(c *gin.Context) {
		token := middleware.EnsureCSRFToken(c)
		c.HTML(200, "login.html", gin.H{"csrf": token})
	})
	r.POST("/login", func(c *gin.Context) {
		cookieToken, err := c.Cookie("csrf_token")
		if err != nil || cookieToken == "" || cookieToken != c.PostForm("_csrf") {
			c.String(http.StatusForbidden, "Invalid CSRF token")
			return
		}

		idStr := c.PostForm("user_id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			token := middleware.EnsureCSRFToken(c)
			c.HTML(200, "login.html", gin.H{"csrf": token})
			return
		}
		middleware.SetUserID(c, id)
		c.Redirect(302, "/notes")
	})

	notes := r.Group("/notes", middleware.AuthRequired(), middleware.CSRFProtection())
	{
		notes.GET("", noteH.Index)
		notes.GET("/create", noteH.Create)
		notes.POST("", noteH.Store)
		notes.GET("/:id/edit", noteH.Edit)
		notes.POST("/:id", noteH.Update)
		notes.POST("/:id/delete", noteH.Destroy)
		notes.POST("/:id/toggle-pin", noteH.TogglePin)
	}

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
