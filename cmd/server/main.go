package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"

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

	r := gin.Default()
	r.Static("/static", "./static")
	tmpl := template.New("")
	tmpl = template.Must(tmpl.ParseGlob("templates/layout/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/notes/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))
	r.SetHTMLTemplate(tmpl)
	view.SetTemplates(tmpl)

	r.GET("/", middleware.AuthRequired(), func(c *gin.Context) {
		c.Redirect(302, "/notes")
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	})
	r.POST("/login", func(c *gin.Context) {
		idStr := c.PostForm("user_id")
		c.SetCookie("user_id", idStr, 3600*24, "/", "", false, true)
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
