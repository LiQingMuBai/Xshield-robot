package domain

import "time"

type EnergyDeploymentLog struct {
	ID               uint      `gorm:"primarykey" json:"id"`
	ChatID           int64     `gorm:"column:chat_id;index:idx_edl_chat_id_created;index:idx_edl_chat_src_created,priority:1" json:"chat_id"`
	TargetAddress    string    `gorm:"column:target_address;size:64;index:idx_edl_addr_created" json:"target_address"`
	Quantity         int       `gorm:"column:quantity;not null;default:1" json:"quantity"`
	Source           string    `gorm:"column:source;size:32;index:idx_edl_chat_src_created,priority:2;index:idx_edl_source_created" json:"source"`
	SourceBundleID   string    `gorm:"column:source_bundle_id;size:64;index:idx_edl_source_bundle" json:"source_bundle_id"`
	SourcePackageID  string    `gorm:"column:source_package_id;size:64;index:idx_edl_source_package" json:"source_package_id"`
	SourceAddressID  string    `gorm:"column:source_address_id;size:64;index:idx_edl_source_addr" json:"source_address_id"`
	BundleTimesAfter int64     `gorm:"column:bundle_times_after;default:0" json:"bundle_times_after"`
	PackageTimesAfter int64    `gorm:"column:package_times_after;default:0" json:"package_times_after"`
	OrderNo          string    `gorm:"column:order_no;size:64;index:idx_edl_order_no" json:"order_no"`
	Status           int       `gorm:"column:status;not null;default:1;index:idx_edl_status_created" json:"status"`
	Provider         string    `gorm:"column:provider;size:16;index:idx_edl_provider_created" json:"provider"`
	ErrorMessage     string    `gorm:"column:error_message;size:512" json:"error_message"`
	CreatedAt        time.Time `gorm:"index:idx_edl_chat_id_created,priority:2;index:idx_edl_addr_created,priority:2;index:idx_edl_chat_src_created,priority:3;index:idx_edl_source_created,priority:2;index:idx_edl_status_created,priority:2;index:idx_edl_provider_created,priority:2"`
}

func (EnergyDeploymentLog) TableName() string {
	return "energy_deployment_log"
}
