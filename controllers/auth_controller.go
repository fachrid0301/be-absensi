package controllers

import (
	"errors"
	"net/http"
	"strings"

	"be-absensi/config"
	"be-absensi/models"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type registerBody struct {
	Nama         string  `json:"nama" binding:"required,min=2,max=100"`
	Email        string  `json:"email" binding:"required,email,max=100"`
	Password     string  `json:"password" binding:"required,min=8,max=72"`
	NimNis       string  `json:"nim_nis" binding:"required,max=50"`
	AsalInstansi string  `json:"asal_instansi" binding:"required,max=150"`
	Jurusan      string  `json:"jurusan" binding:"required,max=100"`
	NoHP         *string `json:"no_hp" binding:"omitempty,max=20"`
}

type loginBody struct {
	NimNis   string `json:"nim_nis" binding:"required,max=50"`
	Password string `json:"password" binding:"required"`
}

// Register buat akun peserta baru (user + data peserta) dalam satu transaksi.
func Register(c *gin.Context) {
	var body registerBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.JSONValidationError(c, formatBindError(err))
		return
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Nama = strings.TrimSpace(body.Nama)
	body.NimNis = strings.TrimSpace(body.NimNis)
	body.AsalInstansi = strings.TrimSpace(body.AsalInstansi)
	body.Jurusan = strings.TrimSpace(body.Jurusan)
	if body.NoHP != nil {
		trimmed := strings.TrimSpace(*body.NoHP)
		if trimmed == "" {
			body.NoHP = nil
		} else {
			body.NoHP = &trimmed
		}
	}

	hash, err := utils.HashPassword(body.Password)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memproses password", nil)
		return
	}

	u := models.User{
		Nama:     body.Nama,
		Email:    body.Email,
		Password: hash,
		Role:     "peserta",
	}
	peserta := models.Peserta{
		NimNis:       body.NimNis,
		AsalInstansi: body.AsalInstansi,
		Jurusan:      body.Jurusan,
		NoHP:         body.NoHP,
		StatusPKL:    "pending",
	}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&u).Error; err != nil {
			return err
		}
		peserta.IDUser = u.IDUser
		return tx.Create(&peserta).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			utils.JSONError(c, http.StatusConflict, "email sudah terdaftar", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal menyimpan registrasi", nil)
		return
	}

	utils.RecordActivity(u.IDUser, "Registrasi akun baru")
	utils.JSONSuccess(c, http.StatusCreated, "registrasi berhasil", gin.H{
		"user":    u.ToPublic(),
		"peserta": peserta,
	})
}

// Login cek nim_nis/password, kembalikan JWT token jika valid.
func Login(c *gin.Context) {
	var body loginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.JSONValidationError(c, formatBindError(err))
		return
	}

	body.NimNis = strings.TrimSpace(body.NimNis)

	var u models.User
	var peserta models.Peserta
	var err error

	err = config.DB.Where("nim_nis = ?", body.NimNis).First(&peserta).Error
	if err == nil {
		if err = config.DB.First(&u, peserta.IDUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.JSONError(c, http.StatusUnauthorized, "nim/nis atau password salah", nil)
				return
			}
			utils.JSONError(c, http.StatusInternalServerError, "gagal memuat pengguna", nil)
			return
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Fallback for Admin (whose email will be entered in the NIM/NIS field)
		if err = config.DB.Where("email = ?", body.NimNis).First(&u).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.JSONError(c, http.StatusUnauthorized, "nim/nis atau password salah", nil)
				return
			}
			utils.JSONError(c, http.StatusInternalServerError, "gagal memuat pengguna", nil)
			return
		}
	} else {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat peserta", nil)
		return
	}

	if err := utils.ComparePassword(u.Password, body.Password); err != nil {
		utils.JSONError(c, http.StatusUnauthorized, "nim/nis atau password salah", nil)
		return
	}

	token, err := utils.GenerateToken(u.IDUser, u.Nama, u.Role)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal membuat token", nil)
		return
	}

	utils.RecordActivity(u.IDUser, "Login ke sistem")
	utils.JSONSuccess(c, http.StatusOK, "login berhasil", gin.H{
		"token": token,
		"user":  u.ToPublic(),
	})
}
