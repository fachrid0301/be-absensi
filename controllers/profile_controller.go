package controllers

import (
	"net/http"

	"be-absensi/backend/middleware"
	"be-absensi/backend/utils"

	"github.com/gin-gonic/gin"
)

func Profile(c *gin.Context) {
	v, ok := c.Get(middleware.ContextUserKey)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}
	claims, ok := v.(*utils.Claims)
	if !ok {
		utils.JSONError(c, http.StatusInternalServerError, "klaim tidak valid", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "profil pengguna", gin.H{
		"id_user": claims.IDUser,
		"nama":    claims.Nama,
		"role":    claims.Role,
	})
}
