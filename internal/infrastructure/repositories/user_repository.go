package repositories

import (
	"context"

	_ "github.com/go-sql-driver/mysql"

	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}
func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
func (r *UserRepository) UpdateBackupChat(ctx context.Context, backup string, associates int64) error {
	query := "UPDATE tg_users SET backup_chat_id = ?  WHERE associates = ?"
	tx := r.db.Exec(query, backup, associates)
	return tx.Error
}

func (r *UserRepository) CreateWithContext(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) Create(user domain.User) error {

	query := "INSERT INTO tg_users (user_id, username,amount,tron_amount,tron_address, eth_address,eth_amount, associates) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	tx := r.db.Exec(query, user.UserID, user.Username, user.Amount, user.TronAmount, user.TronAddress, user.EthAddress, user.EthAmount, user.Associates)

	return tx.Error
}

func (r *UserRepository) UpdateUsernameByChatID(username string, chatID int64) error {
	query := "UPDATE tg_users SET username = ? WHERE associates = ?"
	tx := r.db.Exec(query, username, chatID)
	return tx.Error
}

func (r *UserRepository) Update(user domain.User) error {
	query := "UPDATE tg_users SET associates = $1, tron_amount = $2 WHERE username = $3"
	tx := r.db.Exec(query, user.Associates, user.TronAmount, user.Username)
	return tx.Error
}

func (r *UserRepository) UpdateAddress(user domain.User) error {
	query := "UPDATE tg_users SET address = ? , private_key = ?  WHERE id = ?"
	tx := r.db.Exec(query, user.Address, user.Key, user.Id)
	return tx.Error
}

func (r *UserRepository) UpdateTimes(times uint64, username string) error {
	query := "UPDATE tg_users SET times = ?  WHERE username = ?"
	tx := r.db.Exec(query, times, username)
	return tx.Error
}
func (r *UserRepository) UpdateBundleTimes(bundleTimes int64, chatID int64) error {
	query := "UPDATE tg_users SET bundle_times = ?  WHERE associates = ?"
	tx := r.db.Exec(query, bundleTimes, chatID)
	return tx.Error
}

func (r *UserRepository) UpdateSmartTransactionTimes(bundleTimes int64, chatID int64) error {
	query := "UPDATE tg_users SET st_times = ?  WHERE associates = ?"
	tx := r.db.Exec(query, bundleTimes, chatID)
	return tx.Error
}

func (r *UserRepository) UpdateTRXBalance(trxAmount string, chatID int64) error {
	query := "UPDATE tg_users SET tron_amount = ?  WHERE associates = ?"
	tx := r.db.Exec(query, trxAmount, chatID)
	return tx.Error
}

func (r *UserRepository) UpdateUSDTBalance(amount string, chatID int64) error {
	query := "UPDATE tg_users SET amount = ?  WHERE associates = ?"
	tx := r.db.Exec(query, amount, chatID)
	return tx.Error
}

func (r *UserRepository) UpdateDispatchTimesByChatID(times uint64, chatID int64) error {
	query := "UPDATE tg_users SET times = ?  WHERE associates = ?"
	tx := r.db.Exec(query, times, chatID)
	return tx.Error
}

//associates VARCHAR(255),
//amount VARCHAR(255) ,
//tron_amount VARCHAR(255),
//tron_address VARCHAR(50),
//eth_address VARCHAR(50),
//eth_amount VARCHAR(255),

func (r *UserRepository) GetByUsername(username string) (domain.User, error) {

	userRecord := domain.User{}

	err := r.db.Where(" username=?", username).First(&userRecord).Error

	return userRecord, err
}
func (r *UserRepository) GetByChatID(chatID int64) (domain.User, error) {
	userRecord := domain.User{}

	err := r.db.Where(" associates=?", chatID).First(&userRecord).Error

	return userRecord, err
}
func (r *UserRepository) GetByChatIDString(chatID string) (domain.User, error) {
	userRecord := domain.User{}

	err := r.db.Where(" associates=?", chatID).First(&userRecord).Error

	return userRecord, err
}
func (r *UserRepository) UpdateLang(lang string, chatID int64) error {
	query := "UPDATE tg_users SET lang = ? WHERE associates = ?"
	tx := r.db.Exec(query, lang, chatID)
	return tx.Error
}

func (r *UserRepository) ListActiveAddresses() ([]domain.User, error) {
	query := `SELECT address,associates
    FROM 
      sys_address  where disable=0 ;
    `
	var addresses []domain.User
	r.db.Select(&addresses, query)
	return addresses, nil
}
func (r *UserRepository) DisableTronAddress(address string) error {
	query := "UPDATE sys_address SET disable = 1 WHERE address = ?"
	tx := r.db.Exec(query, address)
	return tx.Error
}

func (r *UserRepository) BindChatID(associates string, username string) error {
	query := "UPDATE tg_users SET associates = ? WHERE username = ?"
	tx := r.db.Exec(query, associates, username)
	return tx.Error
}

func (r *UserRepository) BindTRXAddress(address string, username string) error {
	query := "UPDATE tg_users SET tron_address = ? WHERE username = ?"
	tx := r.db.Exec(query, address, username)
	return tx.Error
}

func (r *UserRepository) BindETHAddress(address string, username string) error {
	query := "UPDATE tg_users SET eth_address = ? WHERE username = ?"
	tx := r.db.Exec(query, address, username)
	return tx.Error
}

func (r *UserRepository) ListNotifiableTRXAddresses() ([]domain.User, error) {
	query := `SELECT t.username,t.tron_address,t.associates
    FROM
        tg_users t
    LEFT JOIN
        sys_address s ON t.tron_address = s.address

    WHERE s.disable = 0
    `
	var addresses []domain.User
	r.db.Select(&addresses, query)
	return addresses, nil
}
func (r *UserRepository) ListNotifiableETHAddresses() ([]domain.User, error) {
	query := `SELECT t.username,t.eth_address,t.associates
    FROM
        tg_users t
    LEFT JOIN
        sys_address s ON t.eth_address = s.address

    WHERE s.disable = 0
    `
	var addresses []domain.User
	r.db.Select(&addresses, query)
	return addresses, nil
}
