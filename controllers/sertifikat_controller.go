package controllers

import (
	"errors"
	"net/http"
	"strings"

	"be-absensi/config"
	"be-absensi/middleware"
	"be-absensi/models"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequestSertifikat(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	var peserta models.Peserta
	if err := config.DB.Where("id_user = ?", claims.IDUser).First(&peserta).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusForbidden, "data peserta tidak ditemukan", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat data peserta", nil)
		return
	}

	if peserta.StatusPKL != "selesai" {
		utils.JSONError(c, http.StatusForbidden, "sertifikat hanya bisa diminta setelah PKL selesai", nil)
		return
	}

	var pending int64
	config.DB.Model(&models.Sertifikat{}).
		Where("id_peserta = ? AND status = ?", peserta.IDPeserta, "pending").
		Count(&pending)
	if pending > 0 {
		utils.JSONError(c, http.StatusConflict, "anda masih memiliki permintaan sertifikat yang menunggu", nil)
		return
	}

	var sudahDiberikan int64
	config.DB.Model(&models.Sertifikat{}).
		Where("id_peserta = ? AND status = ?", peserta.IDPeserta, "diberikan").
		Count(&sudahDiberikan)
	if sudahDiberikan > 0 {
		utils.JSONError(c, http.StatusConflict, "sertifikat sudah pernah diberikan", nil)
		return
	}

	catatan := strings.TrimSpace(c.PostForm("catatan"))
	var catatanPtr *string
	if catatan != "" {
		catatanPtr = &catatan
	}

	s := models.Sertifikat{
		IDPeserta:      peserta.IDPeserta,
		IDUser:         claims.IDUser,
		Status:         "pending",
		Catatan:        catatanPtr,
		TanggalRequest: utils.TodayDate(),
	}

	if err := config.DB.Create(&s).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal mengirim permintaan sertifikat", nil)
		return
	}

	_ = config.DB.Preload("Peserta").Preload("User").First(&s, s.IDSertifikat)
	utils.JSONSuccess(c, http.StatusCreated, "permintaan sertifikat berhasil dikirim", s)
}

func ListSertifikat(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	q := config.DB.Preload("Peserta").Preload("User").Order("id_sertifikat DESC")
	if claims.Role == "peserta" {
		q = q.Where("id_user = ?", claims.IDUser)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}

	var list []models.Sertifikat
	if err := q.Find(&list).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat permintaan sertifikat", nil)
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "daftar permintaan sertifikat", list)
}

func GetSertifikat(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var s models.Sertifikat
	q := config.DB.Preload("Peserta").Preload("User")
	if claims.Role == "peserta" {
		q = q.Where("id_user = ?", claims.IDUser)
	}
	if err := q.First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusNotFound, "permintaan sertifikat tidak ditemukan", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat permintaan sertifikat", nil)
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "detail permintaan sertifikat", s)
}

func VerifikasiSertifikat(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	status := strings.TrimSpace(c.PostForm("status"))
	if status != "diberikan" && status != "ditolak" {
		utils.JSONError(c, http.StatusBadRequest, "status harus diberikan atau ditolak", nil)
		return
	}

	var s models.Sertifikat
	if err := config.DB.First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusNotFound, "permintaan sertifikat tidak ditemukan", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat permintaan sertifikat", nil)
		return
	}

	if s.Status != "pending" {
		utils.JSONError(c, http.StatusBadRequest, "permintaan sudah diproses", nil)
		return
	}

	catatan := strings.TrimSpace(c.PostForm("catatan"))
	if catatan != "" {
		s.Catatan = &catatan
	}

	if status == "diberikan" {
		filePath, err := utils.SaveUploadedFile(c, "file_sertifikat", "sertifikat", utils.AllowedDocExt(), utils.MaxPDFSize)
		if err != nil {
			utils.JSONError(c, http.StatusBadRequest, "File sertifikat: "+err.Error(), nil)
			return
		}
		s.FileSertifikat = &filePath
		now := utils.TodayDate()
		s.TanggalDiberikan = &now
	}

	s.Status = status
	if err := config.DB.Save(&s).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memperbarui permintaan sertifikat", nil)
		return
	}

	_ = config.DB.Preload("Peserta").Preload("User").First(&s, s.IDSertifikat)
	msg := "permintaan sertifikat ditolak"
	if status == "diberikan" {
		msg = "sertifikat berhasil diberikan"
	}
	utils.JSONSuccess(c, http.StatusOK, msg, s)
}
