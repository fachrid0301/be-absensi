package controllers

import (
	"net/http"

	"be-absensi/config"
	"be-absensi/middleware"
	"be-absensi/models"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
)

// ListLogAktivitas mengambil log aktivitas untuk user yang sedang login.
// Admin bisa lihat semua, peserta hanya lihat milik sendiri.
func ListLogAktivitas(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	q := config.DB.Model(&models.LogAktivitas{}).Preload("User").Order("id_log DESC")

	if claims.Role != "admin" {
		q = q.Where("id_user = ?", claims.IDUser)
	}

	var list []models.LogAktivitas
	if err := q.Limit(100).Find(&list).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat log aktivitas", nil)
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "log aktivitas", list)
}
