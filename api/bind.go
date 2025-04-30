package api

import (
	"invite-code-service/dao"
	"invite-code-service/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReqBind struct {
	UserAddress string `json:"user_address"`
	InviteCode  string `json:"invite_code"`
}

type RspBind struct {
	InviteCode string `json:"invite_code"`
}

// @Summary bind user address and invite code
// @Description bind
// @Tags v1
// @Accept json
// @Produce json
// @Param param body ReqBind true "bind"
// @Success 200 {object} utils.Rsp{data=RspBind}
// @Router /v1/invite/bind [post]
func (h *Handler) HandlePostBind(c *gin.Context) {
	req := ReqBind{}
	err := c.Bind(&req)
	if err != nil {
		utils.Err(c, codeParamErr, err.Error())
		logrus.Errorf("bind err %s", err)
		return
	}
	if len(req.UserAddress) == 0 {
		utils.Err(c, codeParamErr, "")
		return
	}

	_, err = dao.GetInviteCodeByUser(h.db, req.UserAddress)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			utils.Err(c, codeInternalErr, err.Error())
			logrus.Errorf("GetInviteCodeByUser err %s", err)
			return
		}
		// pass
	} else {
		utils.Err(c, codeUserAlreadyBoundErr, "")
		return
	}

	var inviteCode *dao.InviteCode
	// bind direct
	if len(req.InviteCode) > 0 {
		inviteCode, err := dao.GetInviteCode(h.db, req.InviteCode)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				utils.Err(c, codeInternalErr, err.Error())
				logrus.Errorf("GetInviteCode err %s", err)
				return
			}

			utils.Err(c, codeInviteCodeNotExistErr, err.Error())
			return
		} else {
			if inviteCode.BindTime != 0 {
				utils.Err(c, codeInviteCodeAlreadyBoundErr, "")
				return
			}

			if inviteCode.CodeType != dao.DirectInviteCode {
				utils.Err(c, codeInviteCodeTypeNotMatchErr, "")
				return
			}
			// pass
		}
	} else {
		// bind task
		inviteCode, err = dao.GetAvailableTaskInviteCode(h.db)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				utils.Err(c, codeInternalErr, err.Error())
				logrus.Errorf("GetAvailableTaskInviteCode err %s", err)
				return
			}

			utils.Err(c, codeInviteCodeNotEnoughErr, err.Error())
			logrus.Errorf("GetAvailableTaskInviteCode err %s", err)
			return
		}
		// pass
	}

	inviteCode.UserAddress = req.UserAddress
	inviteCode.BindTime = uint64(time.Now().Unix())

	err = dao.UpOrInInviteCode(h.db, inviteCode)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("UpOrInInviteCode err %s", err)
		return
	}

	logrus.WithField("inviteCode", inviteCode).Info("bind  success")

	utils.Ok(c, RspBind{
		InviteCode: inviteCode.InviteCode,
	})
}
