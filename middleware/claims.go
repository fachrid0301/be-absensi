package middleware

import (
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
)

// GetClaims ambil data user login dari context (setelah JWTAuth).
func GetClaims(c *gin.Context) (*utils.Claims, bool) {
	v, ok := c.Get(ContextUserKey)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*utils.Claims)
	return claims, ok
}
