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
	csrfHeader     = "X-CSRF-Token"
	csrfKeyLen     = 32
	csrfCtxKey     = "csrf_token"
)

func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {
			cookieToken, err := c.Cookie(csrfCookieName)
			if err != nil || cookieToken == "" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch"})
				return
			}

			requestToken := c.PostForm(csrfFormField)
			if requestToken == "" {
				requestToken = c.GetHeader(csrfHeader)
			}

			if requestToken == "" || cookieToken != requestToken {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch"})
				return
			}
		}

		token := ensureCSRFToken(c)
		c.Set(csrfCtxKey, token)

		c.Next()
	}
}

func GetCSRFToken(c *gin.Context) string {
	v, _ := c.Get(csrfCtxKey)
	token, _ := v.(string)
	return token
}

func ensureCSRFToken(c *gin.Context) string {
	if token, err := c.Cookie(csrfCookieName); err == nil && token != "" {
		return token
	}
	token := generateCSRFToken()
	c.SetCookie(csrfCookieName, token, 3600, "/", "", false, true)
	return token
}

func generateCSRFToken() string {
	b := make([]byte, csrfKeyLen)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic("failed to generate CSRF token: " + err.Error())
	}
	return hex.EncodeToString(b)
}
