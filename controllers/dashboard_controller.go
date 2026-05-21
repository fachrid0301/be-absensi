package controllers

import (
	"net/http"

	"be-absensi/config"
	"be-absensi/models"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
)

type dashboardData struct {
	TotalPeserta         int64 `json:"total_peserta"`
	HadirHariIni         int64 `json:"hadir_hari_ini"`
	TelatHariIni         int64 `json:"telat_hari_ini"`
	PendingPendaftaran   int64 `json:"pending_pendaftaran"`
}

func AdminDashboard(c *gin.Context) {
	today := utils.TodayDate().Format("2006-01-02")

	var totalPeserta int64
	config.DB.Model(&models.Peserta{}).Count(&totalPeserta)

	var hadirHariIni int64
	config.DB.Model(&models.Absensi{}).
		Where("tanggal = ? AND status = ?", today, "hadir").
		Count(&hadirHariIni)

	var telatHariIni int64
	config.DB.Model(&models.Absensi{}).
		Where("tanggal = ? AND status = ?", today, "telat").
		Count(&telatHariIni)

	var pendingPendaftaran int64
	config.DB.Model(&models.Pendaftaran{}).
		Where("status = ?", "pending").
		Count(&pendingPendaftaran)

	utils.JSONSuccess(c, http.StatusOK, "dashboard admin", dashboardData{
		TotalPeserta:       totalPeserta,
		HadirHariIni:       hadirHariIni,
		TelatHariIni:       telatHariIni,
		PendingPendaftaran: pendingPendaftaran,
	})
}
