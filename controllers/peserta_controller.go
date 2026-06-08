package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"be-absensi/config"
	"be-absensi/models"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type pesertaBody struct {
	IDUser       uint    `json:"id_user" binding:"required"`
	NimNis       string  `json:"nim_nis" binding:"required,max=50"`
	AsalInstansi string  `json:"asal_instansi" binding:"required,max=150"`
	Jurusan      string  `json:"jurusan" binding:"required,max=100"`
	NoHP         *string `json:"no_hp" binding:"omitempty,max=20"`
	StatusPKL    string  `json:"status_pkl" binding:"omitempty,oneof=pending diterima ditolak selesai"`
}

func ListPeserta(c *gin.Context) {
	var list []models.Peserta
	q := config.DB.Preload("User")
	if err := q.Order("id_peserta DESC").Find(&list).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat data peserta", nil)
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "daftar peserta", list)
}

func GetPeserta(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var p models.Peserta
	if err := config.DB.Preload("User").First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusNotFound, "peserta tidak ditemukan", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat peserta", nil)
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "detail peserta", p)
}

func CreatePeserta(c *gin.Context) {
	var body pesertaBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.JSONValidationError(c, formatBindError(err))
		return
	}

	var user models.User
	if err := config.DB.First(&user, body.IDUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusBadRequest, "id_user tidak ditemukan", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memvalidasi user", nil)
		return
	}

	status := "pending"
	if body.StatusPKL != "" {
		status = body.StatusPKL
	}

	p := models.Peserta{
		IDUser:       body.IDUser,
		NimNis:       strings.TrimSpace(body.NimNis),
		AsalInstansi: strings.TrimSpace(body.AsalInstansi),
		Jurusan:      strings.TrimSpace(body.Jurusan),
		NoHP:         body.NoHP,
		StatusPKL:    status,
	}

	if err := config.DB.Create(&p).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal menyimpan peserta", nil)
		return
	}
	_ = config.DB.Preload("User").First(&p, p.IDPeserta)
	utils.JSONSuccess(c, http.StatusCreated, "peserta berhasil ditambahkan", p)
}

func UpdatePeserta(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var p models.Peserta
	if err := config.DB.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusNotFound, "peserta tidak ditemukan", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat peserta", nil)
		return
	}

	var body pesertaUpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.JSONValidationError(c, formatBindError(err))
		return
	}

	if body.NimNis != "" {
		p.NimNis = strings.TrimSpace(body.NimNis)
	}
	if body.AsalInstansi != "" {
		p.AsalInstansi = strings.TrimSpace(body.AsalInstansi)
	}
	if body.Jurusan != "" {
		p.Jurusan = strings.TrimSpace(body.Jurusan)
	}
	if body.NoHP != nil {
		p.NoHP = body.NoHP
	}
	if body.StatusPKL != "" {
		p.StatusPKL = body.StatusPKL
		// Sinkronkan status pada tabel pendaftaran jika sudah ada
		var pendaftar models.Pendaftaran
		if err := config.DB.Where("id_user = ?", p.IDUser).First(&pendaftar).Error; err == nil {
			pendaftar.Status = p.StatusPKL
			_ = config.DB.Save(&pendaftar)
		}
	}
	if body.IDUser != 0 && body.IDUser != p.IDUser {
		var user models.User
		if err := config.DB.First(&user, body.IDUser).Error; err != nil {
			utils.JSONError(c, http.StatusBadRequest, "id_user tidak ditemukan", nil)
			return
		}
		p.IDUser = body.IDUser
	}

	if err := config.DB.Save(&p).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memperbarui peserta", nil)
		return
	}
	_ = config.DB.Preload("User").First(&p, p.IDPeserta)
	utils.JSONSuccess(c, http.StatusOK, "peserta berhasil diperbarui", p)
}

func DeletePeserta(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	res := config.DB.Delete(&models.Peserta{}, id)
	if res.Error != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal menghapus peserta", nil)
		return
	}
	if res.RowsAffected == 0 {
		utils.JSONError(c, http.StatusNotFound, "peserta tidak ditemukan", nil)
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "peserta berhasil dihapus", nil)
}

type pesertaUpdateBody struct {
	IDUser       uint    `json:"id_user"`
	NimNis       string  `json:"nim_nis" binding:"omitempty,max=50"`
	AsalInstansi string  `json:"asal_instansi" binding:"omitempty,max=150"`
	Jurusan      string  `json:"jurusan" binding:"omitempty,max=100"`
	NoHP         *string `json:"no_hp" binding:"omitempty,max=20"`
	StatusPKL    string  `json:"status_pkl" binding:"omitempty,oneof=pending diterima ditolak selesai"`
}

func parseIDParam(c *gin.Context, key string) (uint, error) {
	raw := c.Param(key)
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || n == 0 {
		return 0, errors.New("id tidak valid")
	}
	return uint(n), nil
}
