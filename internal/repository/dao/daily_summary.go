package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

type DailySummaryDAO interface {
	FindByUserIDAndDate(ctx context.Context, userID int64, date time.Time) (DailySummary, error)
	AggregateByUserIDAndDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) (DailySummary, error)
}

type GormDailySummaryDAO struct {
	db *gorm.DB
}

func NewGormDailySummaryDAO(db *gorm.DB) DailySummaryDAO {
	return &GormDailySummaryDAO{
		db: db,
	}
}

func (d *GormDailySummaryDAO) FindByUserIDAndDate(ctx context.Context, userID int64, date time.Time) (DailySummary, error) {
	var summary DailySummary
	err := d.db.WithContext(ctx).
		Model(&DailySummary{}).
		Select([]string{
			"total_duration_seconds",
			"max_swing_speed",
			"total_swings",
			"racket_rotation_count",
			"forehand_clear",
			"backhand_clear",
			"forehand_lift",
			"backhand_lift",
			"forehand_net",
			"backhand_net",
			"forehand_smash",
			"backhand_smash",
			"forehand_drop",
			"backhand_drop",
			"forehand_drive",
			"backhand_drive",
			"pickup_count",
		}).
		Where("user_id = ? AND summary_date = ?", userID, date).
		First(&summary).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return summary, nil
	}
	return summary, err
}

func (d *GormDailySummaryDAO) AggregateByUserIDAndDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) (DailySummary, error) {
	var result DailySummary
	err := d.db.WithContext(ctx).
		Model(&DailySummary{}).
		Select([]string{
			"SUM(total_duration_seconds) as total_duration_seconds",
			"MAX(max_swing_speed) as max_swing_speed",
			"SUM(total_swings) as total_swings",
			"SUM(racket_rotation_count) as racket_rotation_count",
			"SUM(forehand_clear) as forehand_clear",
			"SUM(backhand_clear) as backhand_clear",
			"SUM(forehand_lift) as forehand_lift",
			"SUM(backhand_lift) as backhand_lift",
			"SUM(forehand_net) as forehand_net",
			"SUM(backhand_net) as backhand_net",
			"SUM(forehand_smash) as forehand_smash",
			"SUM(backhand_smash) as backhand_smash",
			"SUM(forehand_drop) as forehand_drop",
			"SUM(backhand_drop) as backhand_drop",
			"SUM(forehand_drive) as forehand_drive",
			"SUM(backhand_drive) as backhand_drive",
			"SUM(pickup_count) as pickup_count",
		}).
		Where("user_id = ? AND summary_date BETWEEN ? AND ?", userID, startDate, endDate).
		Scan(&result).Error
	if err != nil {
		return result, err
	}
	return result, err
}

type DailySummary struct {
	ID                   int64     `gorm:"column:id;primaryKey;autoIncrement"` // 主键
	UserID               int64     `gorm:"column:user_id"`                     // 用户ID
	SummaryDate          time.Time `gorm:"column:summary_date;type:date"`      // 汇总日期（格式为 yyyy-MM-dd）
	TotalDurationSeconds int       `gorm:"column:total_duration_seconds"`      // 训练总时长（秒）
	MaxSwingSpeed        int       `gorm:"column:max_swing_speed"`             // 最大挥拍速度
	TotalSwings          int       `gorm:"column:total_swings"`                // 总挥拍次数
	RacketRotationCount  int       `gorm:"column:racket_rotation_count"`       // 转球拍次数

	ForehandClear int `gorm:"column:forehand_clear"` // 正手高远球
	BackhandClear int `gorm:"column:backhand_clear"` // 反手高远球
	ForehandLift  int `gorm:"column:forehand_lift"`  // 正手挑球
	BackhandLift  int `gorm:"column:backhand_lift"`  // 反手挑球
	ForehandNet   int `gorm:"column:forehand_net"`   // 正手搓球
	BackhandNet   int `gorm:"column:backhand_net"`   // 反手搓球
	ForehandSmash int `gorm:"column:forehand_smash"` // 正手杀球
	BackhandSmash int `gorm:"column:backhand_smash"` // 反手杀球
	ForehandDrop  int `gorm:"column:forehand_drop"`  // 正手吊球
	BackhandDrop  int `gorm:"column:backhand_drop"`  // 反手吊球
	ForehandDrive int `gorm:"column:forehand_drive"` // 正手抽球
	BackhandDrive int `gorm:"column:backhand_drive"` // 反手抽球
	PickupCount   int `gorm:"column:pickup_count"`   // 捡球次数

	Ctime int64 `gorm:"column:ctime"` // 创建时间（时间戳）
	Utime int64 `gorm:"column:utime"` // 更新时间（时间戳）
}

func (DailySummary) TableName() string {
	return "daily_summary"
}
