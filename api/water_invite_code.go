package api

import (
	"invite-code-service/dao"
	"invite-code-service/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RspWaterInviteCode struct {
	InviteCodes []string `json:"invite_codes"`
}

// @Summary get water invite code
// @Description get water invite code
// @Tags v1
// @Accept json
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspWaterInviteCode}
// @Router /v1/invite/waterInviteCode [get]
func (h *Handler) GetWaterInviteCode(c *gin.Context) {
	rotations, err := dao.GetWaterRotations(h.db)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("GetWaterRotations err %s", err)
		return
	}

	rsp := RspWaterInviteCode{}
	for _, r := range rotations {
		rsp.InviteCodes = append(rsp.InviteCodes, r.InviteCode)
	}

	utils.Ok(c, rsp)

}
