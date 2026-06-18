package models

import (
	"time"
)

type Absensi struct {
	IDAbsensi  uint       `gorm:"column:id_absensi;primaryKey;autoIncrement" json:"id_absensi"`
	IDPeserta  uint       `gorm:"column:id_peserta;not null" json:"id_peserta"`
	Tanggal    time.Time  `gorm:"column:tanggal;type:date;not null" json:"tanggal"`
	JamMasuk   *string    `gorm:"column:jam_masuk;type:time" json:"jam_masuk,omitempty"`
	JamPulang  *string    `gorm:"column:jam_pulang;type:time" json:"jam_pulang,omitempty"`
	Status     string     `gorm:"column:status;type:enum('hadir','telat','tidak hadir');default:hadir" json:"status"`
	Keterangan string     `gorm:"-" json:"keterangan,omitempty"`
	Foto       *string    `gorm:"column:foto;size:255" json:"foto,omitempty"`
	Lokasi     *string    `gorm:"column:lokasi;size:255" json:"lokasi,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Peserta    *Peserta   `gorm:"foreignKey:IDPeserta;references:IDPeserta" json:"peserta,omitempty"`
}

func (Absensi) TableName() string {
	return "absensi"
}

func (a *Absensi) FillKeterangan(jamMasuk string) {
	if jamMasuk == "" {
		jamMasuk = "08:00"
	}
	if a.Status == "telat" {
		a.Keterangan = "Terlambat — absen masuk melewati jam " + jamMasuk
	} else if a.Status == "hadir" {
		a.Keterangan = "Hadir tepat waktu — absen masuk pada atau sebelum jam " + jamMasuk
	} else if a.Status == "tidak hadir" {
		a.Keterangan = "Tidak hadir / Alpha"
	}
}
