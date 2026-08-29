package domain

import "time"

type AddressCounterparty struct {
	Id               int64     `json:"id" form:"id" gorm:"primarykey;column:id;size:20;"`
	Network          string    `json:"network" gorm:"column:network;size:32;index:idx_cp_addr,unique,priority:1;"`
	CounterpartyAddr string    `json:"counterparty_addr" gorm:"column:counterparty_addr;size:128;index:idx_cp_addr,unique,priority:2;index:idx_cp_counterparty;"`
	Label            string    `json:"label" gorm:"column:label;size:256;index:idx_cp_label;"`
	Title            string    `json:"title" gorm:"column:title;size:512;"`
	RiskScore        int       `json:"risk_score" gorm:"column:risk_score;index:idx_cp_risk_score;"`
	Layer            int       `json:"layer" gorm:"column:layer;"`
	Malicious        int       `json:"malicious" gorm:"column:malicious;"`
	Track            string    `json:"track" gorm:"column:track;size:128;"`
	Color            string    `json:"color" gorm:"column:color;size:64;"`
	Dex              int       `json:"dex" gorm:"column:dex;"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at;"`
	UpdatedAt        time.Time `json:"updatedAt" gorm:"column:updated_at;index:idx_cp_updated;"`
}

func (AddressCounterparty) TableName() string {
	return "address_counterparty"
}
