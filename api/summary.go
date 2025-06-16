package api

import (
	"invite-code-service/dao"
	"invite-code-service/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RspSummary struct {
	TotalCodes     uint64 `json:"total_codes"`
	RemainingCodes uint64 `json:"remaining_codes"`
	Tasks          []Task `json:"tasks"`
}

type Task struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Url         string `json:"url"`
}

// @Summary get codes info and zealy task
// @Description get codes info and zealy task
// @Tags v1
// @Accept json
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspSummary}
// @Router /v1/invite/summary [get]
func (h *Handler) GetSummary(c *gin.Context) {
	tasks, err := h.getTasks()
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("getTasks err %s", err)
		return
	}

	stats, err := dao.GetAllInviteCodeStats(h.db)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("GetTaskInviteCodeStats err %s", err)
		return
	}

	utils.Ok(c, RspSummary{
		TotalCodes:     uint64(stats.TotalCodes),
		RemainingCodes: uint64(stats.RemainCodes),
		Tasks:          tasks,
	})

}
