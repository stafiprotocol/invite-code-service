package api

import (
	"invite-code-service/dao"
	"invite-code-service/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RspUserStatus struct {
	Bound      bool   `json:"bound"`
	InviteCode string `json:"invite_code"`
	BondAt     uint64 `json:"bond_at"`
	CodeType   uint8  `json:"code_type"`
}

// @Summary get user info
// @Description get user info (rsp.status "80000":Success "80001":ParamErr "80002":InternalErr)
// @Tags v1
// @Accept json
// @Produce json
// @Param address query string true "address"
// @Success 200 {object} utils.Rsp{data=RspUserStatus}
// @Router /v1/invite/userStatus [get]
func (h *Handler) GetUserStatus(c *gin.Context) {
	address := c.Query("address")
	if len(address) == 0 {
		utils.Ok(c, "success", RspUserStatus{})
		return
	}

	codeInfo, err := dao.GetInviteCodeByUser(h.db, address)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			utils.Err(c, codeInternalErr, err.Error())
			return
		}
		utils.Ok(c, "success", RspUserStatus{
			Bound:      false,
			InviteCode: "",
			BondAt:     0,
			CodeType:   0,
		})
		return
	}

	utils.Ok(c, "success", RspUserStatus{
		Bound:      true,
		InviteCode: codeInfo.InviteCode,
		BondAt:     codeInfo.BindTime,
		CodeType:   codeInfo.CodeType,
	})

}
