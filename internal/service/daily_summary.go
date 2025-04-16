package service

import (
	"badminton-backend/internal/domain"
	"badminton-backend/internal/repository"
	"context"
	"time"
)

type DailySummaryService interface {
	GetByDate(ctx context.Context, biz string, userID int64, date time.Time) (domain.DailySummary, error)
	GetByDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) (domain.DailySummary, error)
}

type dailySummaryService struct {
	repo repository.DailySummaryRepository
}

func NewDailySummaryService(repo repository.DailySummaryRepository) DailySummaryService {
	return &dailySummaryService{
		repo: repo,
	}
}

func (s *dailySummaryService) GetByDate(ctx context.Context, biz string, userID int64, date time.Time) (domain.DailySummary, error) {
	return s.repo.FindByUserIDAndDate(ctx, biz, userID, date)
}

func (s *dailySummaryService) GetByDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) (domain.DailySummary, error) {
	return s.repo.FindByUserIDAndDateRange(ctx, userID, startDate, endDate)
}
