package repository

import (
	"badminton-backend/internal/domain"
	"badminton-backend/internal/repository/cache"
	"badminton-backend/internal/repository/dao"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type DailySummaryRepository interface {
	FindByUserIDAndDate(ctx context.Context, biz string, userID int64, date time.Time) (domain.DailySummary, error)
	FindByUserIDAndDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) (domain.DailySummary, error)
}

type dailySummaryRepository struct {
	dao   dao.DailySummaryDAO
	cache cache.DailySummaryCache
}

func NewDailySummaryRepository(dao dao.DailySummaryDAO, cache cache.DailySummaryCache) DailySummaryRepository {
	return &dailySummaryRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *dailySummaryRepository) FindByUserIDAndDate(ctx context.Context, biz string, userID int64, date time.Time) (domain.DailySummary, error) {
	ds, err := r.cache.Get(ctx, biz, userID, date)
	switch {
	case err == nil:
		return ds, err
	case errors.Is(err, redis.Nil):
		daoSummary, err := r.dao.FindByUserIDAndDate(ctx, userID, date)
		if err != nil {
			return domain.DailySummary{}, err
		}
		domainSummary := r.entityToDomain(daoSummary)
		_ = r.cache.Set(ctx, biz, userID, date, domainSummary)
		return domainSummary, nil
	default:
		return domain.DailySummary{}, err
	}
}

func (r *dailySummaryRepository) FindByUserIDAndDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) (domain.DailySummary, error) {
	aggSummary, err := r.dao.AggregateByUserIDAndDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return domain.DailySummary{}, err
	}
	aggDomainSummary := r.entityToDomain(aggSummary)
	return aggDomainSummary, nil
}

func (r *dailySummaryRepository) domainToEntity(d domain.DailySummary) dao.DailySummary {
	return dao.DailySummary{
		ID:                   d.ID,
		UserID:               d.UserID,
		SummaryDate:          d.Date,
		TotalDurationSeconds: d.Duration,
		MaxSwingSpeed:        d.MaxSpeed,
		TotalSwings:          d.TotalSwings,
		RacketRotationCount:  d.Rotation,

		ForehandClear: d.ForehandClear,
		BackhandClear: d.BackhandClear,
		ForehandLift:  d.ForehandLift,
		BackhandLift:  d.BackhandLift,
		ForehandNet:   d.ForehandNet,
		BackhandNet:   d.BackhandNet,
		ForehandSmash: d.ForehandSmash,
		BackhandSmash: d.BackhandSmash,
		ForehandDrop:  d.ForehandDrop,
		BackhandDrop:  d.BackhandDrop,
		ForehandDrive: d.ForehandDrive,
		BackhandDrive: d.BackhandDrive,
		PickupCount:   d.PickupCount,

		Ctime: d.CreatedAt.Unix(),
		Utime: d.UpdatedAt.Unix(),
	}
}

func (r *dailySummaryRepository) entityToDomain(ds dao.DailySummary) domain.DailySummary {
	return domain.DailySummary{
		ID:            ds.ID,
		UserID:        ds.UserID,
		Date:          ds.SummaryDate,
		Duration:      ds.TotalDurationSeconds,
		MaxSpeed:      ds.MaxSwingSpeed,
		TotalSwings:   ds.TotalSwings,
		Rotation:      ds.RacketRotationCount,
		ForehandClear: ds.ForehandClear,
		BackhandClear: ds.BackhandClear,
		ForehandLift:  ds.ForehandLift,
		BackhandLift:  ds.BackhandLift,
		ForehandNet:   ds.ForehandNet,
		BackhandNet:   ds.BackhandNet,
		ForehandSmash: ds.ForehandSmash,
		BackhandSmash: ds.BackhandSmash,
		ForehandDrop:  ds.ForehandDrop,
		BackhandDrop:  ds.BackhandDrop,
		ForehandDrive: ds.ForehandDrive,
		BackhandDrive: ds.BackhandDrive,
		PickupCount:   ds.PickupCount,
		CreatedAt:     time.Unix(ds.Ctime, 0),
		UpdatedAt:     time.Unix(ds.Utime, 0),
	}
}
