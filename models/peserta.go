package models

import "time"

type Peserta struct {
	IDPeserta    uint      `gorm:"column:id_peserta;primaryKey;autoIncrement" json:"id_peserta"`
	IDUser       uint      `gorm:"column:id_user;not null" json:"id_user"`
	NimNis       string    `gorm:"column:nim_nis;size:50;not null" json:"nim_nis"`
	AsalInstansi string    `gorm:"column:asal_instansi;size:150;not null" json:"asal_instansi"`
	Jurusan      string    `gorm:"column:jurusan;size:100;not null" json:"jurusan"`
	NoHP         *string   `gorm:"column:no_hp;size:20" json:"no_hp,omitempty"`
	StatusPKL    string    `gorm:"column:status_pkl;type:enum('pending','diterima','ditolak','selesai');default:pending" json:"status_pkl"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	User         *User     `gorm:"foreignKey:IDUser;references:IDUser" json:"user,omitempty"`
}

func (Peserta) TableName() string {
	return "peserta"
}
