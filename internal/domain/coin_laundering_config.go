package domain

type CoinLaunderingConfig struct {
	Id     int64  `json:"id" form:"id" gorm:"primarykey;column:id;size:20;"` //id字段
	Status int64  `json:"status" form:"status" gorm:"column:status;"`        //   `db:"status"`
	Name   string `json:"name" form:"name" gorm:"column:name;"`              // `db:"name"`
	Amount string `json:"amount" form:"amount" gorm:"column:amount;"`        // `db:"amount"`
}

// TableName ronUsers表 RonUsers自定义表名 coin_laundering_config
func (CoinLaunderingConfig) TableName() string {
	return "coin_laundering_config"
}
