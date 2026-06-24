package controllers

import (
	"net/http"
	"strings"

	"be-absensi/config"
	"be-absensi/models"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
)

// GetJadwal ambil jam masuk/pulang; default 08:00–17:00 jika belum diset.
func GetJadwal(c *gin.Context) {
	var j models.Jadwal
	if err := config.DB.First(&j).Error; err != nil {
		// If not found, return default
		j = models.Jadwal{
			ID:        1,
			JamMasuk:  "08:00",
			JamPulang: "17:00",
		}
	}
	utils.JSONSuccess(c, http.StatusOK, "jadwal absensi berhasil dimuat", j)
}

// UpdateJadwal ubah jam masuk/pulang absensi (format HH:MM).
func UpdateJadwal(c *gin.Context) {
	var req struct {
		JamMasuk  string `json:"jam_masuk"`
		JamPulang string `json:"jam_pulang"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "request tidak valid", nil)
		return
	}

	req.JamMasuk = strings.TrimSpace(req.JamMasuk)
	req.JamPulang = strings.TrimSpace(req.JamPulang)

	if req.JamMasuk == "" || req.JamPulang == "" {
		utils.JSONError(c, http.StatusBadRequest, "jam masuk dan jam pulang wajib diisi", nil)
		return
	}

	// Simple validation for HH:MM format
	if len(req.JamMasuk) != 5 || len(req.JamPulang) != 5 || !strings.Contains(req.JamMasuk, ":") || !strings.Contains(req.JamPulang, ":") {
		utils.JSONError(c, http.StatusBadRequest, "format jam harus HH:MM (contoh: 08:00)", nil)
		return
	}

	var j models.Jadwal
	err := config.DB.First(&j).Error
	if err != nil {
		j = models.Jadwal{
			ID:        1,
			JamMasuk:  req.JamMasuk,
			JamPulang: req.JamPulang,
		}
		if err := config.DB.Create(&j).Error; err != nil {
			utils.JSONError(c, http.StatusInternalServerError, "gagal menyimpan jadwal: "+err.Error(), nil)
			return
		}
	} else {
		j.JamMasuk = req.JamMasuk
		j.JamPulang = req.JamPulang
		if err := config.DB.Save(&j).Error; err != nil {
			utils.JSONError(c, http.StatusInternalServerError, "gagal memperbarui jadwal: "+err.Error(), nil)
			return
		}
	}

	utils.JSONSuccess(c, http.StatusOK, "jadwal absensi berhasil diperbarui", j)
}
