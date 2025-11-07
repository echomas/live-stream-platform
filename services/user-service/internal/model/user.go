package model

import "time"

type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string    `gorm:"type:varchar(50);not null" json:"nickname"`
	Gender       int       `gorm:"type:tinyint;default:0" json:"gender"` // 0-未知 1-男性 2-女性
	Avatar       string    `gorm:"type:varchar(255)" json:"avatar"`
	Status       int       `gorm:"type:tinyint;default:1;index" json:"status"` // 0-禁用 1-正常
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
