package repositories

import "gorm.io/gorm"

type SysDictionariesRepo struct {
	db *gorm.DB
}

func NewSysDictionariesRepo(db *gorm.DB) *SysDictionariesRepo {
	return &SysDictionariesRepo{
		db: db,
	}
}

func (r *SysDictionariesRepo) GetDictionary(key string) (string, error) {
	var dict string
	err := r.db.Raw("SELECT description FROM sys_dictionaries where name ='" + key + "'").Scan(&dict).Error
	return dict, err
}

func (r *SysDictionariesRepo) GetReceiveAddress(agent string) (string, error) {
	var dict string
	err := r.db.Raw("SELECT address FROM sys_users where username ='" + agent + "'").Scan(&dict).Error
	return dict, err
}

func (r *SysDictionariesRepo) GetDepositAddress(agent string) (string, error) {
	var dict string
	err := r.db.Raw("SELECT deposit_address FROM sys_users where username ='" + agent + "'").Scan(&dict).Error
	return dict, err
}
func (r *SysDictionariesRepo) GetDictionaryDetail(label string) (string, error) {
	var dict string
	err := r.db.Raw("SELECT value FROM sys_dictionary_details where label ='" + label + "'").Scan(&dict).Error
	return dict, err
}
