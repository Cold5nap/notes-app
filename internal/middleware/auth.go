package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var CookieSecure bool

func SetCookieSecure(secure bool) {
	CookieSecure = secure
}

const (
	SessionUserID = "user_id"
)

// AuthRequired проверяет, что пользователь авторизован.
// В тестовом задании user_id хранится в cookie (имитация сессии).
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		if userID == 0 {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Set(SessionUserID, userID)
		c.Next()
	}
}

// GetUserID возвращает user_id из контекста.
func GetUserID(c *gin.Context) int {
	v, _ := c.Get(SessionUserID)
	id, _ := v.(int)
	return id
}

// getUserID читает user_id из cookie.
// Для тестового задания: если куки нет — пользователь не авторизован.
// В реальном проекте здесь была бы проверка сессионной JWT/куки.
func getUserID(c *gin.Context) int {
	cookie, err := c.Cookie("user_id")
	if err != nil {
		return 0
	}
	id, err := strconv.Atoi(cookie)
	if err != nil {
		return 0
	}
	return id
}

// SetUserID устанавливает cookie user_id (для имитации логина).
func SetUserID(c *gin.Context, userID int) {
	c.SetCookie("user_id", strconv.Itoa(userID), 3600*24, "/", "", CookieSecure, true)
}
