package models

import "time"

type User struct {
	IDUser    uint      `gorm:"column:id_user;primaryKey;autoIncrement" json:"id_user"`
	Nama      string    `gorm:"column:nama;size:100;not null" json:"nama"`
	Email     string    `gorm:"column:email;size:100;not null;uniqueIndex" json:"email"`
	Password  string    `gorm:"column:password;size:255;not null" json:"-"`
	Role      string    `gorm:"column:role;type:enum('admin','peserta');not null;default:peserta" json:"role"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (User) TableName() string {
	return "users"
}

type UserPublic struct {
	IDUser    uint      `json:"id_user"`
	Nama      string    `json:"nama"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) ToPublic() UserPublic {
	return UserPublic{
		IDUser:    u.IDUser,
		Nama:      u.Nama,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
