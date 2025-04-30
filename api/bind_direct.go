package api

import (
	"invite-code-service/dao"
	"invite-code-service/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReqBindDirect struct {
	UserAddress string `json:"user_address"`
	InviteCode  string `json:"invite_code"`
}

// @Summary bind direct
// @Description bind direct
// @Tags v1
// @Accept json
// @Produce json
// @Param param body ReqBindDirect true "bind direct req params"
// @Success 200 {object} utils.Rsp{}
// @Router /v1/invite/bindDirect [post]
func (h *Handler) HandlePostBindDirect(c *gin.Context) {
	req := ReqBindDirect{}
	err := c.Bind(&req)
	if err != nil {
		utils.Err(c, codeParamErr, err.Error())
		logrus.Errorf("bind err %s", err)
		return
	}
	if len(req.UserAddress) == 0 || len(req.InviteCode) == 0 {
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

	inviteCode.UserAddress = req.UserAddress
	inviteCode.BindTime = uint64(time.Now().Unix())

	err = dao.UpOrInInviteCode(h.db, inviteCode)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("UpOrInInviteCode err %s", err)
		return
	}

	logrus.WithField("inviteCode", inviteCode).Info("bind direct success")

	utils.Ok(c, struct{}{})
}
