package config

import (
	"log"

	"be-absensi/models"
)

// EnsureSchema menambahkan kolom/tabel yang belum ada tanpa mengubah tipe kolom int yang sudah ada.
func EnsureSchema() {
	migrator := DB.Migrator()

	if !migrator.HasColumn(&models.Sertifikat{}, "file_berkas") {
		if err := migrator.AddColumn(&models.Sertifikat{}, "FileBerkas"); err != nil {
			log.Printf("gagal menambah kolom file_berkas: %v", err)
		} else {
			log.Println("kolom file_berkas ditambahkan ke tabel sertifikat")
		}
	}
}
