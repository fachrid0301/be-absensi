package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringSlice digunakan untuk menyimpan array string sebagai JSON di database.
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("cannot scan type %T into StringSlice", value)
	}
	return json.Unmarshal(bytes, s)
}

type Sertifikat struct {
	IDSertifikat     uint         `gorm:"column:id_sertifikat;type:int;primaryKey;autoIncrement" json:"id_sertifikat"`
	IDPeserta        uint         `gorm:"column:id_peserta;type:int;not null" json:"id_peserta"`
	IDUser           uint         `gorm:"column:id_user;type:int;not null" json:"id_user"`
	Status           string       `gorm:"column:status;type:enum('pending','diberikan','ditolak');default:pending" json:"status"`
	FileSertifikat   *string      `gorm:"column:file_sertifikat;size:255" json:"file_sertifikat,omitempty"`
	FileBerkas       *StringSlice `gorm:"column:file_berkas;type:text" json:"file_berkas,omitempty"`
	Catatan          *string      `gorm:"column:catatan;type:text" json:"catatan,omitempty"`
	TanggalRequest   time.Time    `gorm:"column:tanggal_request;type:date;not null" json:"tanggal_request"`
	TanggalDiberikan *time.Time   `gorm:"column:tanggal_diberikan;type:date" json:"tanggal_diberikan,omitempty"`
	Peserta          *Peserta     `gorm:"foreignKey:IDPeserta;references:IDPeserta" json:"peserta,omitempty"`
	User             *User        `gorm:"foreignKey:IDUser;references:IDUser" json:"user,omitempty"`
}

func (Sertifikat) TableName() string {
	return "sertifikat"
}
