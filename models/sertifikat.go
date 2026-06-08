package models

import "time"

type Sertifikat struct {
	IDSertifikat     uint       `gorm:"column:id_sertifikat;primaryKey;autoIncrement" json:"id_sertifikat"`
	IDPeserta        uint       `gorm:"column:id_peserta;not null" json:"id_peserta"`
	IDUser           uint       `gorm:"column:id_user;not null" json:"id_user"`
	Status           string     `gorm:"column:status;type:enum('pending','diberikan','ditolak');default:pending" json:"status"`
	FileSertifikat   *string    `gorm:"column:file_sertifikat;size:255" json:"file_sertifikat,omitempty"`
	Catatan          *string    `gorm:"column:catatan;type:text" json:"catatan,omitempty"`
	TanggalRequest   time.Time  `gorm:"column:tanggal_request;type:date;not null" json:"tanggal_request"`
	TanggalDiberikan *time.Time `gorm:"column:tanggal_diberikan;type:date" json:"tanggal_diberikan,omitempty"`
	Peserta          *Peserta   `gorm:"foreignKey:IDPeserta;references:IDPeserta" json:"peserta,omitempty"`
	User             *User      `gorm:"foreignKey:IDUser;references:IDUser" json:"user,omitempty"`
}

func (Sertifikat) TableName() string {
	return "sertifikat"
}
