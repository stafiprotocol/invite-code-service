package api

import (
	"invite-code-service/pkg/utils"

	"github.com/gin-gonic/gin"
)

type RspTasks struct {
	Tasks []Task `json:"tasks"`
}

type Task struct {
	Id          string `json:"id"`
	Description string `json:"description"`
}

// @Summary get tasks
// @Description get tasks
// @Tags v1
// @Accept json
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspTasks}
// @Router /v1/invite/tasks [get]
func (h *Handler) GetTasks(c *gin.Context) {

	utils.Ok(c, RspTasks{
		Tasks: []Task{},
	})

}
