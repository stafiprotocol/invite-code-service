package api

import (
	"invite-code-service/dao"
	"invite-code-service/pkg/utils"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RspDroplets struct {
	Droplets []Droplet `json:"droplets"`
}

type Droplet struct {
	TotalCount     uint64 `json:"total_count"`
	AvailableCount uint64 `json:"available_count"`
	Round          uint8  `json:"round"`
	InviteCode     string `json:"invite_code"`
}

// @Summary get droplets
// @Description get droplets
// @Tags v1
// @Accept json
// @Produce json
// @Param droplet query string false "droplet"
// @Success 200 {object} utils.Rsp{data=RspDroplets}
// @Router /v1/invite/droplets [get]
func (h *Handler) GetDroplets(c *gin.Context) {
	droplet := c.Query("droplet")
	dropletCodes, err := dao.GetLatestDropletCodesWithStatus(h.db)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("GetWaterRotations err %s", err)
		return
	}

	rsp := ConvertToRspDroplets(dropletCodes)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(rsp.Droplets), func(i, j int) {
		rsp.Droplets[i], rsp.Droplets[j] = rsp.Droplets[j], rsp.Droplets[i]
	})

	if droplet == "sp" {
		// [3,5]
		count := len(rsp.Droplets)
		if count > 5 {
			count = rand.Intn(3) + 3
		} else if count >= 3 {
			count = rand.Intn(count-2) + 3
		}
		rsp.Droplets = rsp.Droplets[:count]
	} else {
		// [0,3] 80%
		// [4,5] 20%
		var finalCount int
		p := rand.Float64()

		if p < 0.8 {
			max := min(3, len(rsp.Droplets))

			finalCount = rand.Intn(max + 1)
		} else {
			minCount := min(4, len(rsp.Droplets))
			maxCount := min(5, len(rsp.Droplets))

			finalCount = rand.Intn(maxCount-minCount+1) + minCount

		}
		rsp.Droplets = rsp.Droplets[:finalCount]
	}

	utils.Ok(c, rsp)

}

func ConvertToRspDroplets(data []*dao.DropletCodeWithStatus) RspDroplets {
	// key: round + dropletIndex
	type key struct {
		Round        uint8
		DropletIndex uint8
	}

	groupMap := make(map[key][]*dao.DropletCodeWithStatus)

	for _, item := range data {
		k := key{Round: item.Round, DropletIndex: item.DropletIndex}
		groupMap[k] = append(groupMap[k], item)
	}

	var droplets []Droplet
	for k, list := range groupMap {
		var availableCount uint64
		var selectedCode string

		for _, d := range list {
			if !d.Used {
				availableCount++
				if selectedCode == "" {
					selectedCode = d.InviteCode
				}
			}
		}

		droplets = append(droplets, Droplet{
			TotalCount:     uint64(len(list)),
			AvailableCount: availableCount,
			Round:          k.Round,
			InviteCode:     selectedCode,
		})
	}

	return RspDroplets{Droplets: droplets}
}
