package middleware

import (
	"be-absensi/backend/utils"

	"github.com/gin-gonic/gin"
)

func GetClaims(c *gin.Context) (*utils.Claims, bool) {
	v, ok := c.Get(ContextUserKey)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*utils.Claims)
	return claims, ok
}
