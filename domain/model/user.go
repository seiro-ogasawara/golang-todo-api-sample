package model

type User struct {
	UserID   string `gorm:"primaryKey"`
	Password string `gorm:"not null"`
}

func (User) TableName() string {
	return "users"
}
