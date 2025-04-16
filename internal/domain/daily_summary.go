package domain

import "time"

type DailySummary struct {
	ID          int64     // 主键，通常内部不需要，用于数据库标识
	UserID      int64     // 用户 ID
	Date        time.Time // 汇总日期
	Duration    int       // 总训练时长（秒）
	MaxSpeed    int       // 最大挥拍速度
	TotalSwings int       // 总挥拍次数
	Rotation    int       // 转球拍次数

	// 击球类型统计
	ForehandClear int
	BackhandClear int
	ForehandLift  int
	BackhandLift  int
	ForehandNet   int
	BackhandNet   int
	ForehandSmash int
	BackhandSmash int
	ForehandDrop  int
	BackhandDrop  int
	ForehandDrive int
	BackhandDrive int
	PickupCount   int

	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}
