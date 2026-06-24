package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"be-absensi/config"
	"be-absensi/middleware"
	"be-absensi/models"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MarkAbsentPeserta dipanggil oleh scheduler setelah jam pulang lewat.
// Fungsi ini membuat record "tidak hadir" untuk setiap peserta aktif
// yang belum memiliki catatan absensi pada tanggal yang diberikan.
func MarkAbsentPeserta(tanggal time.Time, _ string) (int, error) {
	// Ambil semua peserta dengan status diterima
	var pesertaList []models.Peserta
	if err := config.DB.Where("status_pkl = ?", "diterima").Find(&pesertaList).Error; err != nil {
		return 0, err
	}

	count := 0
	status := "tidak hadir"

	for _, p := range pesertaList {
		// Cek apakah sudah ada absensi hari ini
		var existing models.Absensi
		err := config.DB.
			Where("id_peserta = ? AND tanggal = ?", p.IDPeserta, tanggal).
			First(&existing).Error

		if err == nil {
			// Sudah absen, lewati
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Error lain, lewati peserta ini
			continue
		}

		// Buat record tidak hadir
		abs := models.Absensi{
			IDPeserta: p.IDPeserta,
			Tanggal:   tanggal,
			Status:    status,
		}
		if createErr := config.DB.Create(&abs).Error; createErr != nil {
			continue
		}
		count++
	}

	return count, nil
}

// findPesertaByUser ambil data peserta berdasarkan id_user.
func findPesertaByUser(idUser uint) (*models.Peserta, error) {
	var p models.Peserta
	err := config.DB.Where("id_user = ?", idUser).First(&p).Error
	return &p, err
}

// AbsensiMasuk catat kehadiran masuk: cek status PKL, jadwal, upload foto + lokasi.
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

	if peserta.StatusPKL != "diterima" {
		msg := "Pendaftaran Anda belum disetujui. Anda tidak dapat melakukan absensi."
		if peserta.StatusPKL == "pending" {
			msg = "Pendaftaran Anda masih ditinjau. Silakan tunggu persetujuan admin."
		} else if peserta.StatusPKL == "ditolak" {
			msg = "Pendaftaran Anda ditolak. Anda tidak diperkenankan melakukan absensi."
		}
		utils.JSONError(c, http.StatusForbidden, msg, nil)
		return
	}

	today := utils.TodayDate()

	var existing models.Absensi
	err = config.DB.
		Where("id_peserta = ? AND tanggal = ?", peserta.IDPeserta, today).
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

	var jadwal models.Jadwal
	if err := config.DB.First(&jadwal).Error; err != nil {
		jadwal.JamMasuk = "08:00"
		jadwal.JamPulang = "17:00"
	}

	now := time.Now()
	jam := now.Format("15:04:05")

	isLate := false
	var hour, minute int
	_, errScan := fmt.Sscanf(jadwal.JamMasuk, "%d:%d", &hour, &minute)
	if errScan == nil {
		deadline := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if now.After(deadline) {
			isLate = true
		}
	} else {
		if now.Hour() > 8 || (now.Hour() == 8 && now.Minute() > 0) {
			isLate = true
		}
	}

	if isLate {
		utils.JSONError(c, http.StatusBadRequest, "batas waktu absensi masuk telah lewat ("+jadwal.JamMasuk+")", nil)
		return
	}

	abs := models.Absensi{
		IDPeserta: peserta.IDPeserta,
		Tanggal:   today,
		JamMasuk:  &jam,
		Status:    "hadir",
		Foto:      &fotoPath,
		Lokasi:    &lokasi,
	}

	if err := config.DB.Create(&abs).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal menyimpan absensi masuk", nil)
		return
	}

	abs.FillKeterangan(jadwal.JamMasuk)
	utils.JSONSuccess(c, http.StatusCreated, "absensi masuk berhasil", abs)
}

// AbsensiPulang catat pulang setelah absen masuk, sesuai jam pulang jadwal.
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

	if peserta.StatusPKL != "diterima" {
		msg := "Pendaftaran Anda belum disetujui. Anda tidak dapat melakukan absensi."
		if peserta.StatusPKL == "pending" {
			msg = "Pendaftaran Anda masih ditinjau. Silakan tunggu persetujuan admin."
		} else if peserta.StatusPKL == "ditolak" {
			msg = "Pendaftaran Anda ditolak. Anda tidak diperkenankan melakukan absensi."
		}
		utils.JSONError(c, http.StatusForbidden, msg, nil)
		return
	}

	today := utils.TodayDate()
	var abs models.Absensi
	err = config.DB.Where("id_peserta = ? AND tanggal = ?", peserta.IDPeserta, today).First(&abs).Error
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

	var jadwal models.Jadwal
	if err := config.DB.First(&jadwal).Error; err != nil {
		jadwal.JamMasuk = "08:00"
		jadwal.JamPulang = "17:00"
	}

	now := time.Now()
	var pHour, pMinute int
	_, errScan := fmt.Sscanf(jadwal.JamPulang, "%d:%d", &pHour, &pMinute)
	if errScan == nil {
		pulangTime := time.Date(now.Year(), now.Month(), now.Day(), pHour, pMinute, 0, 0, now.Location())
		if now.Before(pulangTime) {
			utils.JSONError(c, http.StatusBadRequest, "belum memasuki jam pulang ("+jadwal.JamPulang+")", nil)
			return
		}
	} else {
		if now.Hour() < 17 {
			utils.JSONError(c, http.StatusBadRequest, "belum memasuki jam pulang (17:00)", nil)
			return
		}
	}

	jam := now.Format("15:04:05")
	abs.JamPulang = &jam

	if err := config.DB.Save(&abs).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal menyimpan absensi pulang", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "absensi pulang berhasil", abs)
}

// AbsensiHistory riwayat absensi; peserta lihat milik sendiri, admin bisa filter.
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

	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse("2006-01-02", fromStr); err == nil {
			q = q.Where("tanggal >= ?", from)
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse("2006-01-02", toStr); err == nil {
			q = q.Where("tanggal <= ?", to)
		}
	}

	var list []models.Absensi
	if err := q.Find(&list).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat riwayat absensi", nil)
		return
	}
	var jadwal models.Jadwal
	if err := config.DB.First(&jadwal).Error; err != nil {
		jadwal.JamMasuk = "08:00"
	}
	for i := range list {
		list[i].FillKeterangan(jadwal.JamMasuk)
	}
	utils.JSONSuccess(c, http.StatusOK, "riwayat absensi", list)
}

type absensiStatusBody struct {
	Status string `json:"status" binding:"required"`
}

// UpdateAbsensiStatus admin ubah status absensi: hadir, telat, atau tidak hadir.
func UpdateAbsensiStatus(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var body absensiStatusBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.JSONValidationError(c, formatBindError(err))
		return
	}

	switch body.Status {
	case "hadir", "telat", "tidak hadir":
	default:
		utils.JSONError(c, http.StatusBadRequest, "status harus hadir, telat, atau tidak hadir", nil)
		return
	}

	var abs models.Absensi
	if err := config.DB.First(&abs, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusNotFound, "absensi tidak ditemukan", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat absensi", nil)
		return
	}

	abs.Status = body.Status
	if body.Status == "tidak hadir" {
		abs.JamMasuk = nil
		abs.JamPulang = nil
	}

	if err := config.DB.Save(&abs).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memperbarui status absensi", nil)
		return
	}

	var jadwal models.Jadwal
	if err := config.DB.First(&jadwal).Error; err != nil {
		jadwal.JamMasuk = "08:00"
	}
	_ = config.DB.Preload("Peserta").Preload("Peserta.User").First(&abs, abs.IDAbsensi)
	abs.FillKeterangan(jadwal.JamMasuk)
	utils.JSONSuccess(c, http.StatusOK, "status absensi berhasil diperbarui", abs)
}
