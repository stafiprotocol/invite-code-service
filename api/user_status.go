package api

import (
	"invite-code-service/dao"
	"invite-code-service/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RspUserStatus struct {
	Bound      bool   `json:"bound"`
	InviteCode string `json:"invite_code"`
	Tasks      []Task `json:"tasks"`
}

// @Summary get user status
// @Description get user status
// @Tags v1
// @Accept json
// @Produce json
// @Param address query string true "address"
// @Success 200 {object} utils.Rsp{data=RspUserStatus}
// @Router /v1/invite/userStatus [get]
func (h *Handler) GetUserStatus(c *gin.Context) {
	address := c.Query("address")
	if len(address) == 0 {
		utils.Ok(c, RspUserStatus{})
		return
	}
	address = strings.ToLower(address)

	codeInfo, err := dao.GetInviteCodeByUserAddress(h.db, address)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			utils.Err(c, codeInternalErr, err.Error())
			logrus.Errorf("bind err %s", err)
			return
		}

		userTasks, err := h.getUserTasks(address)
		if err != nil {
			if err != utils.ErrAddressNotFound {
				utils.Err(c, codeInternalErr, err.Error())
				logrus.Errorf("getUserTasks err %s", err)
				return
			} else {
				userTasks = []Task{}
			}
		}

		utils.Ok(c, RspUserStatus{
			Bound:      false,
			InviteCode: "",
			Tasks:      userTasks,
		})
		return
	}

	rsp := RspUserStatus{
		Bound:      true,
		InviteCode: codeInfo.InviteCode,
		Tasks:      []Task{},
	}
	if codeInfo.CodeType == dao.TaskInviteCode {
		userTasks, err := h.getUserTasks(address)
		if err == nil {
			rsp.Tasks = userTasks
		}
	}

	utils.Ok(c, rsp)
}
