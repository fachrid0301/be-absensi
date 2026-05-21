package controllers

import (
	"errors"
	"net/http"
	"time"

	"be-absensi/config"
	"be-absensi/middleware"
	"be-absensi/models"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func findPesertaByUser(idUser uint) (*models.Peserta, error) {
	var p models.Peserta
	err := config.DB.Where("id_user = ?", idUser).First(&p).Error
	return &p, err
}

func AbsensiMasuk(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	peserta, err := findPesertaByUser(claims.IDUser)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusNotFound, "data peserta tidak ditemukan untuk akun ini", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat peserta", nil)
		return
	}

	today := utils.TodayDate()

	var existing models.Absensi
	err = config.DB.
		Where("id_peserta = ? AND tanggal = ?", peserta.IDPeserta, today.Format("2006-01-02")).
		First(&existing).Error

	if err == nil {
		utils.JSONError(c, http.StatusConflict, "anda sudah absen masuk hari ini", nil)
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memeriksa absensi", nil)
		return
	}

	lokasi := c.PostForm("lokasi")
	if lokasi == "" {
		utils.JSONError(c, http.StatusBadRequest, "lokasi wajib diisi", nil)
		return
	}

	fotoPath, err := utils.SaveUploadedFile(c, "foto", "absensi", utils.AllowedImageExt(), utils.MaxImageSize)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	now := time.Now()
	jam := utils.TimeOnly(now)
	status := utils.AttendanceStatus(now)

	abs := models.Absensi{
		IDPeserta: peserta.IDPeserta,
		Tanggal:   today,
		JamMasuk:  &jam,
		Status:    status,
		Foto:      &fotoPath,
		Lokasi:    &lokasi,
	}

	if err := config.DB.Create(&abs).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal menyimpan absensi masuk", nil)
		return
	}

	abs.FillKeterangan()
	utils.JSONSuccess(c, http.StatusCreated, "absensi masuk berhasil", abs)
}

func AbsensiPulang(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	peserta, err := findPesertaByUser(claims.IDUser)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusNotFound, "data peserta tidak ditemukan untuk akun ini", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat peserta", nil)
		return
	}

	today := utils.TodayDate()
	var abs models.Absensi
	err = config.DB.Where("id_peserta = ? AND tanggal = ?", peserta.IDPeserta, today.Format("2006-01-02")).First(&abs).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		utils.JSONError(c, http.StatusBadRequest, "belum absen masuk hari ini", nil)
		return
	}
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat absensi", nil)
		return
	}
	if abs.JamPulang != nil {
		utils.JSONError(c, http.StatusConflict, "anda sudah absen pulang hari ini", nil)
		return
	}

	lokasi := c.PostForm("lokasi")
	if lokasi != "" {
		abs.Lokasi = &lokasi
	}

	now := time.Now()
	jam := utils.TimeOnly(now)
	abs.JamPulang = &jam

	if err := config.DB.Save(&abs).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal menyimpan absensi pulang", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "absensi pulang berhasil", abs)
}

func AbsensiHistory(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	q := config.DB.Model(&models.Absensi{}).Preload("Peserta").Preload("Peserta.User").Order("tanggal DESC, id_absensi DESC")

	if claims.Role == "peserta" {
		peserta, err := findPesertaByUser(claims.IDUser)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.JSONSuccess(c, http.StatusOK, "riwayat absensi", []models.Absensi{})
				return
			}
			utils.JSONError(c, http.StatusInternalServerError, "gagal memuat peserta", nil)
			return
		}
		q = q.Where("id_peserta = ?", peserta.IDPeserta)
	} else if idPeserta := c.Query("id_peserta"); idPeserta != "" {
		q = q.Where("id_peserta = ?", idPeserta)
	}

	if from := c.Query("from"); from != "" {
		q = q.Where("tanggal >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		q = q.Where("tanggal <= ?", to)
	}

	var list []models.Absensi
	if err := q.Find(&list).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat riwayat absensi", nil)
		return
	}
	for i := range list {
		list[i].FillKeterangan()
	}
	utils.JSONSuccess(c, http.StatusOK, "riwayat absensi", list)
}
