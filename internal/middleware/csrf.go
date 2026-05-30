package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	csrfCookieName = "csrf_token"
	csrfFormField  = "_csrf"
	csrfKeyLen     = 32
)

func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {
			cookieToken, err := c.Cookie(csrfCookieName)
			formToken := c.PostForm(csrfFormField)

			if err != nil || cookieToken == "" || formToken == "" || cookieToken != formToken {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch"})
				return
			}
		}

		if _, err := c.Cookie(csrfCookieName); err != nil {
			token := generateCSRFToken()
			c.SetCookie(csrfCookieName, token, 3600, "/", "", false, true)
		}

		c.Next()
	}
}

func generateCSRFToken() string {
	b := make([]byte, csrfKeyLen)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic("failed to generate CSRF token: " + err.Error())
	}
	return hex.EncodeToString(b)
}
