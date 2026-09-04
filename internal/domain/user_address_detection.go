package domain

import "time"

type UserAddressDetection struct {
	Id          int64     `json:"id" form:"id" gorm:"primarykey;column:id;size:20;"`
	ChatID      int64     `json:"chat_id" form:"chat_id" gorm:"column:chat_id;"`
	Status      int64     `json:"status" form:"status" gorm:"column:status;"`
	Network     string    `json:"network" form:"network" gorm:"column:network;"`
	Address     string    `json:"address" form:"address" gorm:"column:address;"`
	Amount      string    `json:"amount" form:"amount" gorm:"column:amount;"`
	Currency    string    `json:"currency" form:"currency" gorm:"column:currency;size:8;index;"`
	CreatedAt   time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`
	UpdatedAt   time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`
	CreatedDate string    `json:"created_date" `
}

// TableName ronUsers表 RonUsers自定义表名 ron_users
func (UserAddressDetection) TableName() string {
	return "user_address_detection"
}
