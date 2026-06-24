package controllers

import (
	"net/http"

	"be-absensi/middleware"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
)

// Profile kembalikan info user dari JWT (id, nama, role).
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
