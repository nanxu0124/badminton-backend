package service

import (
	"badminton-backend/internal/domain"
	"badminton-backend/internal/repository"
	"badminton-backend/pkg/logger"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("用户名或密码不正确")
)

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	Login(ctx context.Context, account, password string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error
}

// UserService 表示用户相关的业务逻辑服务
type userService struct {
	repo   repository.UserRepository // 引用repository层的UserRepository对象，用于数据访问
	logger logger.Logger
}

// NewUserService 实现 UserService 接口
func NewUserService(repo repository.UserRepository, l logger.Logger) UserService {
	return &userService{
		repo:   repo,
		logger: l,
	}
}

func (svc *userService) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

// FindOrCreate 如果手机号不存在，那么会初始化一个用户
func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 这是一种优化写法
	// 大部分人会命中这个分支
	u, err := svc.repo.FindByPhone(ctx, phone)       // 从数据库中查找用户
	if !errors.Is(err, repository.ErrUserNotFound) { // 如果用户已经存在，则直接返回
		return u, err
	}
	// 如果找不到用户，则执行用户注册操作
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone, // 创建新用户时只需要手机号
	})
	// 注册过程中，如果发生了非手机号码冲突的错误，说明是系统错误
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return domain.User{}, err // 返回错误，表示用户创建失败
	}
	// 如果注册成功或者是重复注册（用户已经存在），从数据库重新查询该手机号的用户
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) Login(ctx context.Context, account, password string) (domain.User, error) {
	u, err := svc.repo.FindByAccount(ctx, account)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, err
}

func (svc *userService) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {

	user.Account = ""
	user.Phone = ""
	user.Password = ""
	return svc.repo.Update(ctx, user)
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return svc.repo.FindById(ctx, id)
}
