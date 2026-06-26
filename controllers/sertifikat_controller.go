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

// RequestSertifikat peserta ajukan berkas; max 1 pending, tidak boleh duplikat.
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

	var pending int64
	config.DB.Model(&models.Sertifikat{}).
		Where("id_peserta = ? AND status = ?", peserta.IDPeserta, "pending").
		Count(&pending)
	if pending > 0 {
		utils.JSONError(c, http.StatusConflict, "anda masih memiliki permintaan berkas yang menunggu", nil)
		return
	}

	var sudahDiberikan int64
	config.DB.Model(&models.Sertifikat{}).
		Where("id_peserta = ? AND status = ?", peserta.IDPeserta, "diberikan").
		Count(&sudahDiberikan)
	if sudahDiberikan > 0 {
		utils.JSONError(c, http.StatusConflict, "berkas sudah pernah diberikan", nil)
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
		utils.JSONError(c, http.StatusInternalServerError, "gagal mengirim permintaan berkas", nil)
		return
	}

	_ = config.DB.Preload("Peserta").Preload("User").First(&s, s.IDSertifikat)
	utils.RecordActivity(claims.IDUser, "Mengajukan permintaan berkas PKL")
	utils.JSONSuccess(c, http.StatusCreated, "permintaan berkas berhasil dikirim", s)
}

// ListSertifikat daftar permintaan; peserta lihat milik sendiri, admin lihat semua.
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
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat permintaan berkas", nil)
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "daftar permintaan berkas", list)
}

// GetSertifikat detail permintaan berkas by ID.
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
			utils.JSONError(c, http.StatusNotFound, "permintaan berkas tidak ditemukan", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat permintaan berkas", nil)
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "detail permintaan berkas", s)
}

// VerifikasiSertifikat admin setujui (upload file) atau tolak permintaan berkas.
// Admin dapat mengupload lebih dari 1 file berkas (PDF).
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
			utils.JSONError(c, http.StatusNotFound, "permintaan berkas tidak ditemukan", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat permintaan berkas", nil)
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
		// Ambil semua file berkas yang diupload (bisa lebih dari 1)
		form, err := c.MultipartForm()
		if err != nil {
			utils.JSONError(c, http.StatusBadRequest, "gagal memproses form data", nil)
			return
		}

		files := form.File["file_berkas"]
		if len(files) == 0 {
			// Fallback ke field lama file_sertifikat untuk kompatibilitas
			files = form.File["file_sertifikat"]
		}

		if len(files) == 0 {
			utils.JSONError(c, http.StatusBadRequest, "minimal 1 file berkas wajib diunggah", nil)
			return
		}

		var berkasFiles models.StringSlice
		for _, fileHeader := range files {
			// Validasi setiap file
			if fileHeader.Size > utils.MaxPDFSize {
				utils.JSONError(c, http.StatusBadRequest, "ukuran file maksimal 10 MB per file", nil)
				return
			}
			ext := strings.ToLower(fileHeader.Filename[strings.LastIndex(fileHeader.Filename, ".")+1:])
			if ext != "pdf" {
				utils.JSONError(c, http.StatusBadRequest, "hanya file PDF yang diizinkan", nil)
				return
			}

			// Simpan file menggunakan SaveMultipartFile
			filePath, saveErr := utils.SaveMultipartFileHeader(fileHeader, "sertifikat")
			if saveErr != nil {
				utils.JSONError(c, http.StatusInternalServerError, "gagal menyimpan file: "+saveErr.Error(), nil)
				return
			}
			berkasFiles = append(berkasFiles, filePath)
		}

		s.FileBerkas = &berkasFiles
		// Juga simpan file pertama ke FileSertifikat untuk kompatibilitas backward
		if len(berkasFiles) > 0 {
			s.FileSertifikat = &berkasFiles[0]
		}
		now := utils.TodayDate()
		s.TanggalDiberikan = &now
	}

	s.Status = status
	if err := config.DB.Save(&s).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memperbarui permintaan berkas", nil)
		return
	}

	_ = config.DB.Preload("Peserta").Preload("User").First(&s, s.IDSertifikat)
	msg := "permintaan berkas ditolak"
	if status == "diberikan" {
		msg = "berkas berhasil diberikan"
		utils.RecordActivity(s.IDUser, "Berkas PKL telah diberikan oleh admin")
	} else {
		utils.RecordActivity(s.IDUser, "Permintaan berkas PKL ditolak oleh admin")
	}
	utils.JSONSuccess(c, http.StatusOK, msg, s)
}
