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

type verifikasiBody struct {
	Status string `json:"status" binding:"required,oneof=diterima ditolak"`
}

func CreatePendaftaran(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	var pending int64
	config.DB.Model(&models.Pendaftaran{}).
		Where("id_user = ? AND status = ?", claims.IDUser, "pending").
		Count(&pending)
	if pending > 0 {
		utils.JSONError(c, http.StatusConflict, "anda masih memiliki pendaftaran pending", nil)
		return
	}

	fileSuratPath, err := utils.SaveUploadedFile(c, "file_surat", "pendaftaran", utils.AllowedDocExt(), utils.MaxPDFSize)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Surat pengantar: "+err.Error(), nil)
		return
	}

	fileCVPath, err := utils.SaveUploadedFile(c, "file_cv", "pendaftaran", utils.AllowedDocExt(), utils.MaxPDFSize)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "CV: "+err.Error(), nil)
		return
	}

	fileLamaranPath, err := utils.SaveUploadedFile(c, "file_surat_lamaran", "pendaftaran", utils.AllowedDocExt(), utils.MaxPDFSize)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Surat lamaran: "+err.Error(), nil)
		return
	}

	p := models.Pendaftaran{
		IDUser:           claims.IDUser,
		FileSurat:        fileSuratPath,
		FileCV:           fileCVPath,
		FileSuratLamaran: fileLamaranPath,
		Status:           "pending",
		TanggalDaftar:    utils.TodayDate(),
	}

	if err := config.DB.Create(&p).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal menyimpan pendaftaran", nil)
		return
	}
	_ = config.DB.Preload("User").First(&p, p.IDPendaftaran)
	utils.JSONSuccess(c, http.StatusCreated, "pendaftaran berhasil dikirim", p)
}

func ListPendaftaran(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	q := config.DB.Preload("User").Order("id_pendaftaran DESC")
	if claims.Role == "peserta" {
		q = q.Where("id_user = ?", claims.IDUser)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}

	var list []models.Pendaftaran
	if err := q.Find(&list).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat pendaftaran", nil)
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "daftar pendaftaran", list)
}

func VerifikasiPendaftaran(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var body verifikasiBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.JSONValidationError(c, formatBindError(err))
		return
	}

	var p models.Pendaftaran
	if err := config.DB.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusNotFound, "pendaftaran tidak ditemukan", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat pendaftaran", nil)
		return
	}

	if p.Status != "pending" {
		utils.JSONError(c, http.StatusBadRequest, "pendaftaran sudah diverifikasi", nil)
		return
	}

	p.Status = strings.TrimSpace(body.Status)
	if err := config.DB.Save(&p).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memperbarui pendaftaran", nil)
		return
	}

	// Sinkronkan status_pkl pada tabel peserta jika sudah ada
	var peserta models.Peserta
	if err := config.DB.Where("id_user = ?", p.IDUser).First(&peserta).Error; err == nil {
		peserta.StatusPKL = p.Status
		_ = config.DB.Save(&peserta)
	}

	_ = config.DB.Preload("User").First(&p, p.IDPendaftaran)
	utils.JSONSuccess(c, http.StatusOK, "pendaftaran berhasil diverifikasi", p)
}
