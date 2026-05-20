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
	Nama     string `json:"nama" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type loginBody struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var body registerBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.JSONValidationError(c, formatBindError(err))
		return
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Nama = strings.TrimSpace(body.Nama)

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

	if err := config.DB.Create(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			utils.JSONError(c, http.StatusConflict, "email sudah terdaftar", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal menyimpan pengguna", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "registrasi berhasil", u.ToPublic())
}

func Login(c *gin.Context) {
	var body loginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.JSONValidationError(c, formatBindError(err))
		return
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))

	var u models.User
	if err := config.DB.Where("email = ?", body.Email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusUnauthorized, "email atau password salah", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat pengguna", nil)
		return
	}

	if err := utils.ComparePassword(u.Password, body.Password); err != nil {
		utils.JSONError(c, http.StatusUnauthorized, "email atau password salah", nil)
		return
	}

	token, err := utils.GenerateToken(u.IDUser, u.Nama, u.Role)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "gagal membuat token", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "login berhasil", gin.H{
		"token": token,
		"user":  u.ToPublic(),
	})
}
