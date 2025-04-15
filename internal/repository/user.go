package repository

import (
	"badminton-backend/internal/domain"
	"badminton-backend/internal/repository/cache"
	"badminton-backend/internal/repository/dao"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrDataNotFound
)

// UserRepository 与数据库交互
type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByAccount(ctx context.Context, account string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	// Update 更新数据，只有非 0 值才会更新
	Update(ctx context.Context, u domain.User) error
}

// CachedUserRepository 实现 UserRepository 接口
type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCachedUserRepository(d dao.UserDAO, c cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   d,
		cache: c,
	}
}

func (ur *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return ur.dao.Insert(ctx, dao.User{
		Account: sql.NullString{
			String: u.Account,
			Valid:  u.Account != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
	})
}

func (ur *CachedUserRepository) FindByPhone(ctx context.Context,
	phone string) (domain.User, error) {
	u, err := ur.dao.FindByPhone(ctx, phone)
	return ur.entityToDomain(u), err
}

func (ur *CachedUserRepository) FindByAccount(ctx context.Context, account string) (domain.User, error) {
	u, err := ur.dao.FindByAccount(ctx, account)
	return ur.entityToDomain(u), err
}

func (ur *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := ur.cache.Get(ctx, id)
	switch {
	case err == nil:
		return u, err
	case errors.Is(err, cache.ErrKeyNotExist):
		ue, err := ur.dao.FindById(ctx, id)
		if err != nil {
			return domain.User{}, err
		}
		u = ur.entityToDomain(ue)
		_ = ur.cache.Set(ctx, u)
		return u, nil
	default:
		return domain.User{}, err
	}
}

func (ur *CachedUserRepository) Update(ctx context.Context, u domain.User) error {
	err := ur.dao.UpdateNonZeroFields(ctx, ur.domainToEntity(u))
	if err != nil {
		return err
	}
	return ur.cache.Delete(ctx, u.Id)
}

func (ur *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Username: sql.NullString{
			String: u.Username,
			Valid:  u.Username != "",
		},
		Account: sql.NullString{
			String: u.Account,
			Valid:  u.Account != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Birthday: sql.NullInt64{
			Int64: u.Birthday.UnixMilli(),
			Valid: !u.Birthday.IsZero(),
		},
		Nickname: sql.NullString{
			String: u.Nickname,
			Valid:  u.Nickname != "",
		},
		AboutMe: sql.NullString{
			String: u.AboutMe,
			Valid:  u.AboutMe != "",
		},
		Password: u.Password,
		HeightCm: u.HeightCM,
		WeightKg: u.WeightKG,
	}
}

func (ur *CachedUserRepository) entityToDomain(ue dao.User) domain.User {
	var birthday time.Time
	if ue.Birthday.Valid {
		birthday = time.UnixMilli(ue.Birthday.Int64)
	}
	return domain.User{
		Id:       ue.Id,
		Username: ue.Username.String,
		Account:  ue.Account.String,
		Password: ue.Password,
		Phone:    ue.Phone.String,
		Nickname: ue.Nickname.String,
		AboutMe:  ue.AboutMe.String,
		Birthday: birthday,
		WeightKG: ue.WeightKg,
		HeightCM: ue.HeightCm,
		Ctime:    time.UnixMilli(ue.Ctime),
	}
}
