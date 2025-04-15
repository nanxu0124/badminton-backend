package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	// ErrUserDuplicate 表示用户邮箱或者手机号冲突错误
	ErrUserDuplicate = errors.New("用户邮箱或者手机号冲突")

	// ErrDataNotFound 通用的数据没找到错误（即Gorm的记录未找到）
	ErrDataNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByAccount(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	UpdateNonZeroFields(ctx context.Context, u User) error
}

// GormUserDAO 与用户数据表交互的所有操作
type GormUserDAO struct {
	db *gorm.DB
}

func NewGormUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDAO{
		db: db,
	}
}

func (ud *GormUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now

	err := ud.db.WithContext(ctx).Create(&u).Error
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		const uniqueIndexErrNo uint16 = 1062 // 唯一索引冲突错误码
		if me.Number == uniqueIndexErrNo {
			// 如果是唯一索引冲突，返回自定义的 ErrUserDuplicate 错误
			return ErrUserDuplicate
		}
	}
	return err
}

func (ud *GormUserDAO) FindByAccount(ctx context.Context, account string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).First(&u, "account = ?", account).Error
	return u, err
}

func (ud *GormUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).First(&u, "phone = ?", phone).Error
	return u, err
}

func (ud *GormUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).First(&u, "id = ?", id).Error
	return u, err
}

func (ud *GormUserDAO) UpdateNonZeroFields(ctx context.Context, u User) error {
	return ud.db.Updates(&u).Error
}

type User struct {
	Id       int64
	Username sql.NullString
	Account  sql.NullString
	Password string
	Phone    sql.NullString
	Nickname sql.NullString
	WeightKg int
	HeightCm int
	AboutMe  sql.NullString
	Birthday sql.NullInt64
	Ctime    int64
	Utime    int64
}
