package api

import (
	"invite-code-service/dao"
	"invite-code-service/pkg/utils"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReqGen struct {
	UserAddress string `json:"user_address"`
	Signature   string `json:"signature"`
	Timestamp   uint64 `json:"timestamp"`
}

type RspGen struct {
	InviteCode string `json:"invite_code"`
}

// @Summary gen invite code
// @Description The exact message format to sign is here:
// @Description https://github.com/stafiprotocol/invite-code-service/blob/main/pkg/utils/signature.go
// @Tags v1
// @Accept json
// @Produce json
// @Param param body ReqGen true "gen"
// @Success 200 {object} utils.Rsp{data=RspGen}
// @Router /v1/invite/genInviteCode [post]
func (h *Handler) HandlePostGenInviteCode(c *gin.Context) {
	req := ReqGen{}
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

	signMessage := utils.BuildGenMessage(req.Timestamp)
	if !utils.VerifySigsEthPersonal(sigBts, signMessage, userAddress) {
		utils.Err(c, codeUserSigVerifyErr, "verify sigs failed")
		logrus.Errorf("VerifySigsEthPersonal failed, user: %s", req.UserAddress)
		return
	}

	// check task
	tasks, err := h.getTasks()
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("getTasks err %s", err)
		return
	}
	userTasks, err := h.getUserTasks(req.UserAddress)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("getUserTasks err %s", err)
		return
	}
	if len(userTasks) == 0 || len(tasks) == 0 || len(userTasks) < len(tasks) {
		utils.Err(c, codeUserTaskVerifyErr, "")
		logrus.Errorf("task not enough, userTasks len: %d, tasks len: %d", len(userTasks), len(tasks))
		return
	}

	// bind task
	inviteCode, err := dao.GetAvailableTaskInviteCode(h.db)
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

	userInfo, err := h.getUserInfo(req.UserAddress)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("getUserInfo err %s", err)
		return
	}

	if len(userInfo.DiscordID) == 0 {
		utils.Err(c, codeInternalErr, "")
		logrus.Errorf("user discord empty err, user: %s", req.UserAddress)
		return
	}

	inviteCode.UserId = &userInfo.ID
	inviteCode.DiscordId = &userInfo.DiscordID
	inviteCode.UserAddress = &req.UserAddress
	inviteCode.BindTime = uint64(time.Now().Unix())

	err = dao.UpOrInInviteCode(h.db, inviteCode)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("UpOrInInviteCode err %s", err)
		return
	}

	logrus.WithFields(logrus.Fields{
		"inviteCode":  inviteCode,
		"userAddress": req.UserAddress,
	}).Info("bind  success")

	utils.Ok(c, RspGen{
		InviteCode: inviteCode.InviteCode,
	})
}
