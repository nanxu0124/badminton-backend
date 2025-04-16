package web

import (
	"badminton-backend/internal/service"
	ijwt "badminton-backend/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	bizDailySummary = "DailySummary"
)

type DailySummaryHandler struct {
	svc service.DailySummaryService
}

func NewDailySummaryHandler(svc service.DailySummaryService) *DailySummaryHandler {
	return &DailySummaryHandler{
		svc: svc,
	}
}

func (h *DailySummaryHandler) RegisterRoutes(server *gin.Engine) {
	v1 := server.Group("/api/v1")
	g := v1.Group("/daily-summary")

	g.POST("/date", h.GetByDate)
	g.POST("/range", h.GetByDateRange)
}

func (h *DailySummaryHandler) GetByDate(ctx *gin.Context) {
	type Req struct {
		Date string `json:"date"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	date, err := time.Parse(time.DateOnly, req.Date)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 14002,
			Msg:  "日期格式不对",
		})
		return
	}

	uc := ctx.MustGet("user").(ijwt.UserClaims)
	summary, err := h.svc.GetByDate(ctx, bizDailySummary, uc.Id, date)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 10000,
		Msg:  "OK",
		Data: summary,
	})
}

func (h *DailySummaryHandler) GetByDateRange(ctx *gin.Context) {
	type Req struct {
		StartDateStr string `json:"start_date"`
		EndDateStr   string `json:"end_date"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	startDate, err := time.Parse(time.DateOnly, req.StartDateStr)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 14002,
			Msg:  "日期格式不对",
		})
		return
	}
	endDate, err := time.Parse(time.DateOnly, req.EndDateStr)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 14002,
			Msg:  "日期格式不对",
		})
		return
	}

	uc := ctx.MustGet("user").(ijwt.UserClaims)
	summaries, err := h.svc.GetByDateRange(ctx, uc.Id, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 10000,
		Msg:  "OK",
		Data: summaries,
	})
}
