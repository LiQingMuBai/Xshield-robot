package domain

import "time"

type CoinLaunderingOrder struct {
	Id          int64     `json:"id" form:"id" gorm:"primarykey;column:id;size:20;"`            //id字段
	Status      int64     `json:"status" form:"status" gorm:"column:status;"`                   //   `db:"status"`
	ChatID      int64     `json:"chat_id" form:"chat_id" gorm:"column:chat_id;"`                //   `db:"user_id"`
	Amount      string    `json:"amount" form:"amount" gorm:"column:amount;"`                   // `db:"amount"`
	OrderNO     string    `json:"order_no" form:"order_no" gorm:"column:order_no;"`             // `db:"times"`
	FromAddress string    `json:"from_address" form:"from_address" gorm:"column:from_address;"` // `db:"times"`
	ToAddress   string    `json:"to_address" form:"to_address" gorm:"column:to_address;"`       // `db:"times"`
	Token       string    `json:"token" form:"token" gorm:"column:token;"`                      // `db:"times"`
	CreatedAt   time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`         //createdAt字段 `db:"create_at"`
	UpdatedAt   time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`         //updatedAt字段`db:"update_at"`
}

// TableName ronUsers表 RonUsers自定义表名 coin_laundering_order
func (CoinLaunderingOrder) TableName() string {
	return "coin_laundering_order"
}
