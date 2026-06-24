package middleware

import (
	"net/http"
	"strings"

	"be-absensi/utils"

	"github.com/gin-gonic/gin"
)

const ContextUserKey = "auth_user" // key untuk simpan JWT claims di gin.Context

// JWTAuth validasi header Bearer token, simpan claims ke context jika valid.
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
