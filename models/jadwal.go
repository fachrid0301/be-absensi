package models

import "time"

type Jadwal struct {
	ID        uint      `gorm:"column:id;primaryKey" json:"id"`
	JamMasuk  string    `gorm:"column:jam_masuk;type:varchar(5);not null;default:'08:00'" json:"jam_masuk"`
	JamPulang string    `gorm:"column:jam_pulang;type:varchar(5);not null;default:'17:00'" json:"jam_pulang"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Jadwal) TableName() string {
	return "jadwal"
}
