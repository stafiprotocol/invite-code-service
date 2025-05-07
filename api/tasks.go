package api

import (
	"invite-code-service/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	tasks, err := h.getTasks()
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("getTasks err %s", err)
		return
	}

	utils.Ok(c, RspTasks{
		Tasks: tasks,
	})

}
