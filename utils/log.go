package utils

import (
	"be-absensi/config"
	"be-absensi/models"
)

// RecordActivity menyimpan log aktivitas ke database untuk user tertentu.
func RecordActivity(idUser uint, aktivitas string) {
	if idUser == 0 {
		return
	}
	log := models.LogAktivitas{
		IDUser:    idUser,
		Aktivitas: aktivitas,
	}
	_ = config.DB.Create(&log)
}
