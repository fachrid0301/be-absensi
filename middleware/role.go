package middleware


import (
	"net/http"

	"be-absensi/utils"

	"github.com/gin-gonic/gin"
)

// AdminOnly batasi endpoint hanya untuk role admin.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
			c.Abort()
			return
		}
		if claims.Role != "admin" {
			utils.JSONError(c, http.StatusForbidden, "akses khusus admin", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

// PesertaOnly batasi endpoint hanya untuk role peserta.
func PesertaOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
			c.Abort()
			return
		}
		if claims.Role != "peserta" {
			utils.JSONError(c, http.StatusForbidden, "akses khusus peserta", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
