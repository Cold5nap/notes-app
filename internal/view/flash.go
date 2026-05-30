package view

import (
	"bytes"
	"html/template"

	"github.com/gin-gonic/gin"
)

var appTmpl *template.Template

func SetTemplates(tmpl *template.Template) {
	appTmpl = tmpl
}

type FlashMessage struct {
	Text string
	Type string
}

const flashCookieName = "flash_message"

func SetFlash(c *gin.Context, text, msgType string) {
	c.SetCookie(flashCookieName, msgType+"|"+text, 0, "/", "", false, true)
}

func GetFlash(c *gin.Context) *FlashMessage {
	cookie, err := c.Cookie(flashCookieName)
	if err != nil {
		return nil
	}

	c.SetCookie(flashCookieName, "", -1, "/", "", false, true)

	if len(cookie) < 2 {
		return nil
	}

	return &FlashMessage{
		Type: string(cookie[0]),
		Text: cookie[2:],
	}
}

func Render(c *gin.Context, contentTemplate string, data any) {
	flash := GetFlash(c)
	token, _ := c.Cookie("csrf_token")

	var buf bytes.Buffer
	err := appTmpl.ExecuteTemplate(&buf, contentTemplate, gin.H{
		"data": data,
		"csrf": token,
	})
	if err != nil {
		c.String(500, "template error: %v", err)
		return
	}

	c.HTML(200, "base.html", gin.H{
		"content": template.HTML(buf.String()),
		"flash":   flash,
		"csrf":    token,
	})
}
