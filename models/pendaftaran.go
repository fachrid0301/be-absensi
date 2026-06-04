package models

import "time"

type Pendaftaran struct {
	IDPendaftaran    uint      `gorm:"column:id_pendaftaran;primaryKey;autoIncrement" json:"id_pendaftaran"`
	IDUser           uint      `gorm:"column:id_user;not null" json:"id_user"`
	FileSurat        string    `gorm:"column:file_surat;size:255;not null" json:"file_surat"`
	FileCV           string    `gorm:"column:file_cv;size:255;not null" json:"file_cv"`
	FileSuratLamaran string    `gorm:"column:file_surat_lamaran;size:255;not null" json:"file_surat_lamaran"`
	Status           string    `gorm:"column:status;type:enum('pending','diterima','ditolak');default:pending" json:"status"`
	TanggalDaftar    time.Time `gorm:"column:tanggal_daftar;type:date;not null" json:"tanggal_daftar"`
	User             *User     `gorm:"foreignKey:IDUser;references:IDUser" json:"user,omitempty"`
}

func (Pendaftaran) TableName() string {
	return "pendaftaran"
}
