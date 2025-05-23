package api

import (
	"errors"
	"invite-code-service/dao"
	"invite-code-service/pkg/utils"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReqBind struct {
	UserAddress string `json:"user_address"`
	DiscordId   string `json:"discord_id"`
	DiscordName string `json:"discord_name"`
	InviteCode  string `json:"invite_code"`
	Signature   string `json:"signature"`
	Timestamp   uint64 `json:"timestamp"`
}

// @Summary bind user address and invite code
// @Description The exact message format to sign is here:
// @Description https://github.com/stafiprotocol/invite-code-service/blob/main/pkg/utils/signature.go
// @Tags v1
// @Accept json
// @Produce json
// @Param param body ReqBind true "bind"
// @Success 200 {object} utils.Rsp{}
// @Router /v1/invite/bind [post]
func (h *Handler) HandlePostBind(c *gin.Context) {
	req := ReqBind{}
	err := c.Bind(&req)
	if err != nil {
		utils.Err(c, codeParamErr, err.Error())
		logrus.Errorf("bind err %s", err)
		return
	}
	if len(req.UserAddress) == 0 || len(req.DiscordId) == 0 || len(req.DiscordName) == 0 || len(req.InviteCode) == 0 || len(req.Signature) == 0 {
		utils.Err(c, codeParamErr, "")
		return
	}
	req.UserAddress = strings.ToLower(req.UserAddress)

	sigBts := common.FromHex(req.Signature)
	userAddress := common.HexToAddress(req.UserAddress)

	_, err = dao.GetInviteCodeByUserAddress(h.db, req.UserAddress)
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

	// check signature
	if !utils.IsValidSignTime(req.Timestamp) {
		utils.Err(c, codeUserSigVerifyErr, "invalid sign time")
		logrus.Errorf("IsValidSignTime failed, user: %s", req.UserAddress)
		return
	}

	signMessage := utils.BuildBindMessage(req.InviteCode, req.DiscordId, req.DiscordName, req.Timestamp)
	if !utils.VerifySigsEthPersonal(sigBts, signMessage, userAddress) {
		utils.Err(c, codeUserSigVerifyErr, "verify sigs failed")
		logrus.Errorf("VerifySigsEthPersonal failed, user: %s", req.UserAddress)
		return
	}

	// bind direct or water invite code
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

		if inviteCode.CodeType != dao.DirectInviteCode && inviteCode.CodeType != dao.WaterInviteCode {
			utils.Err(c, codeInviteCodeTypeNotMatchErr, "")
			return
		}
		// pass
	}

	inviteCode.UserAddress = &req.UserAddress
	inviteCode.DiscordId = &req.DiscordId
	inviteCode.DiscordName = &req.DiscordName
	inviteCode.BindTime = uint64(time.Now().Unix())

	err = dao.CheckBondAndUpdateInviteCode(h.db, inviteCode)
	if err != nil {
		if errors.Is(err, dao.ErrAlreadyBond) {
			utils.Err(c, codeUserAlreadyBoundErr, "")
			return
		}

		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("UpOrInInviteCode err %s", err)
		return
	}

	logrus.WithFields(logrus.Fields{
		"req":        req,
		"inviteCode": inviteCode,
	}).Info("bind  success")

	utils.Ok(c, nil)
}
