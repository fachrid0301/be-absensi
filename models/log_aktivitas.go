package models

import "time"

type LogAktivitas struct {
	IDLog     uint      `gorm:"column:id_log;primaryKey;autoIncrement" json:"id_log"`
	IDUser    uint      `gorm:"column:id_user;not null" json:"id_user"`
	Aktivitas string    `gorm:"column:aktivitas;type:text" json:"aktivitas"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	User      *User     `gorm:"foreignKey:IDUser;references:IDUser" json:"user,omitempty"`
}

func (LogAktivitas) TableName() string {
	return "log_aktivitas"
}
