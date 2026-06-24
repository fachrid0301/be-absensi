package controllers

import (
	"errors"
	"net/http"

	"be-absensi/config"
	"be-absensi/middleware"
	"be-absensi/models"
	"be-absensi/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Profile kembalikan info user; untuk peserta sertakan NIM/NIS dan peminatan.
func Profile(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		utils.JSONError(c, http.StatusUnauthorized, "tidak terautentikasi", nil)
		return
	}

	data := gin.H{
		"id_user": claims.IDUser,
		"nama":    claims.Nama,
		"role":    claims.Role,
	}

	var peserta models.Peserta
	err := config.DB.Where("id_user = ?", claims.IDUser).First(&peserta).Error
	if err == nil {
		data["nim_nis"] = peserta.NimNis
		if peserta.Peminatan != nil {
			data["peminatan"] = *peserta.Peminatan
		} else {
			data["peminatan"] = nil
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.JSONError(c, http.StatusInternalServerError, "gagal memuat profil", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "profil pengguna", data)
}
