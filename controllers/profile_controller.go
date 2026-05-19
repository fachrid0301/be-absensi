package controllers

import (
	"net/http"

	"be-absensi/backend/middleware"
	"be-absensi/backend/utils"

	"github.com/gin-gonic/gin"
)

func Profile(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "profil pengguna", gin.H{
		"id_user": claims.IDUser,
		"nama":    claims.Nama,
		"role":    claims.Role,
	})
}
