package models

import "time"

type Absensi struct {
	IDAbsensi uint      `gorm:"column:id_absensi;primaryKey;autoIncrement" json:"id_absensi"`
	IDPeserta uint      `gorm:"column:id_peserta;not null" json:"id_peserta"`
	Tanggal   time.Time `gorm:"column:tanggal;type:date;not null" json:"tanggal"`
	JamMasuk  *time.Time `gorm:"column:jam_masuk;type:time" json:"jam_masuk,omitempty"`
	JamPulang *time.Time `gorm:"column:jam_pulang;type:time" json:"jam_pulang,omitempty"`
	Status    string    `gorm:"column:status;type:enum('hadir','telat','tidak hadir');default:hadir" json:"status"`
	Foto      *string   `gorm:"column:foto;size:255" json:"foto,omitempty"`
	Lokasi    *string   `gorm:"column:lokasi;size:255" json:"lokasi,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Peserta   *Peserta  `gorm:"foreignKey:IDPeserta;references:IDPeserta" json:"peserta,omitempty"`
}

func (Absensi) TableName() string {
	return "absensi"
}
