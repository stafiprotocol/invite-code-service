package api

import (
	"invite-code-service/api"
	"invite-code-service/dao"
	"invite-code-service/pkg/config"
	"invite-code-service/pkg/db"
	"invite-code-service/pkg/utils"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Task struct {
	cfg *config.ConfigApi

	httpServer *http.Server
	db         *db.WrapDb
}

func NewTask(cfg *config.ConfigApi, dao *db.WrapDb) (*Task, error) {
	s := &Task{
		cfg: cfg,
		db:  dao,
	}

	handler := s.InitHandler()

	s.httpServer = &http.Server{
		Addr:         s.cfg.ListenAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return s, nil
}

func (svr *Task) InitHandler() http.Handler {
	return api.InitRouters(svr.db, svr.cfg)
}

func (svr *Task) ApiServer() {
	logrus.Infof("Gin server start on %s", svr.cfg.ListenAddr)
	err := svr.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logrus.Errorf("Gin server start err: %s", err.Error())
		utils.ShutdownRequestChannel <- struct{}{} //shutdown server
		return
	}
	logrus.Infof("Gin server done on %s", svr.cfg.ListenAddr)
}

func (svr *Task) Start() error {
	directInviteCodeCount, err := dao.GetInviteCodeCount(svr.db, dao.DirectInviteCode)
	if err != nil {
		return err
	}
	taskInviteCodeCount, err := dao.GetInviteCodeCount(svr.db, dao.TaskInviteCode)
	if err != nil {
		return err
	}

	if directInviteCodeCount < int64(svr.cfg.DirectInviteCodeCount) {
		genCount := int64(svr.cfg.DirectInviteCodeCount) - directInviteCodeCount
		logrus.Infof("need generate %d direct invite code", genCount)
		err := svr.genInviteCode(genCount, dao.DirectInviteCode)
		if err != nil {
			return err
		}
		logrus.Infof("generate success")
	}

	if taskInviteCodeCount < int64(svr.cfg.TaskInviteCodeCount) {
		genCount := int64(svr.cfg.TaskInviteCodeCount) - taskInviteCodeCount

		logrus.Infof("need generate %d task invite code", genCount)
		err := svr.genInviteCode(genCount, dao.TaskInviteCode)
		if err != nil {
			return err
		}
		logrus.Infof("generate success")
	}

	utils.SafeGoWithRestart(svr.ApiServer)
	return nil
}

func (svr *Task) genInviteCode(genCount int64, codeType uint8) error {
	for i := int64(0); i < genCount; i++ {
		inviteCode, err := utils.GenerateInviteCode()
		if err != nil {
			return err
		}

		_, err = dao.GetInviteCode(svr.db, inviteCode)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			// pass
		} else {
			i--
			continue
		}

		newInviteCode := dao.InviteCode{
			InviteCode: inviteCode,
			CodeType:   codeType,
		}

		err = dao.UpOrInInviteCode(svr.db, &newInviteCode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svr *Task) Stop() {
	if svr.httpServer != nil {
		err := svr.httpServer.Close()
		if err != nil {
			logrus.Errorf("Problem shutdown Gin server :%s", err.Error())
		}
	}
}
