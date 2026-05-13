package middleware

import (
	"net/http"
	"strings"

	"be-absensi/backend/utils"

	"github.com/gin-gonic/gin"
)

const ContextUserKey = "auth_user"

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			utils.JSONError(c, http.StatusUnauthorized, "header Authorization wajib", nil)
			c.Abort()
			return
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.JSONError(c, http.StatusUnauthorized, "format: Bearer <token>", nil)
			c.Abort()
			return
		}
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.JSONError(c, http.StatusUnauthorized, "token tidak valid atau kadaluarsa", nil)
			c.Abort()
			return
		}
		c.Set(ContextUserKey, claims)
		c.Next()
	}
}
