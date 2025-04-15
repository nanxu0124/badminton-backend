package service

import (
	"badminton-backend/internal/repository"
	"badminton-backend/internal/service/sms"
	"badminton-backend/pkg/logger"
	"context"
	"errors"
	"fmt"
	"math/rand"
)

var (
	ErrCodeSendTooMany = repository.ErrCodeSendTooMany
)

const codeTplId = "2044585" // 短信模板 ID，用于发送验证码短信

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type SMSCodeService struct {
	sms    sms.Service
	repo   repository.CodeRepository
	logger logger.Logger
}

func NewSMSCodeService(svc sms.Service, repo repository.CodeRepository, l logger.Logger) CodeService {
	return &SMSCodeService{
		sms:    svc,
		repo:   repo,
		logger: l,
	}
}

func (c *SMSCodeService) Send(ctx context.Context, biz string, phone string) error {
	code := c.generate()

	err := c.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}

	err = c.sms.Send(ctx, codeTplId, []string{code}, phone)
	if err != nil {
		c.logger.Warn("发送验证码短信失败: ", logger.Field{
			Key:   "SMSCodeService",
			Value: err.Error(),
		})
	}
	return err
}

func (c *SMSCodeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	ok, err := c.repo.Verify(ctx, biz, phone, inputCode)
	if errors.Is(err, repository.ErrCodeVerifyTooManyTimes) {
		c.logger.Error("验证次数超过限制: ", logger.Field{
			Key:   "SMSCodeService",
			Value: err.Error(),
		})
		return false, nil
	}
	return ok, err
}

// generate 生成一个随机的 6 位验证码
// 使用随机数生成一个介于 0 到 999999 之间的验证码
func (c *SMSCodeService) generate() string {
	num := rand.Intn(999999) // 生成随机数
	// 将随机数格式化为 6 位字符串，前面补零
	return fmt.Sprintf("%06d", num)
}
