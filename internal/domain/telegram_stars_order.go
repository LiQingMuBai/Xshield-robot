package domain

import "time"

type TelegramStarsOrder struct {
	Id         int64     `json:"id" form:"id" gorm:"primarykey;column:id;size:20;"`         //id字段
	ChatID     int64     `json:"chat_id" form:"chat_id" gorm:"column:chat_id;"`             //   `db:"user_id"`
	Status     int64     `json:"status" form:"status" gorm:"column:status;"`                //   `db:"status"`
	TGUsername string    `json:"tg_username" form:"tg_username" gorm:"column:tg_username;"` // `db:"name"`
	OrderNO    string    `json:"order_no" form:"order_no" gorm:"column:order_no;"`          // `db:"name"`
	Stars      string    `json:"stars" form:"stars" gorm:"column:stars;"`                   // `db:"stars"`
	Amount     string    `json:"amount" form:"amount" gorm:"column:amount;"`                // `db:"amount"`
	CreatedAt  time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`      //createdAt字段 `db:"create_at"`
	UpdatedAt  time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`      //updatedAt字段`db:"update_at"`
}

// TableName telegram_Stars_order RonUsers自定义表名 telegram_Stars_order
func (TelegramStarsOrder) TableName() string {
	return "telegram_stars_order"
}
