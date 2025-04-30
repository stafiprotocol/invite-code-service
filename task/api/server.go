package api

import (
	"invite-code-service/api"
	"invite-code-service/pkg/config"
	"invite-code-service/pkg/db"
	"invite-code-service/pkg/utils"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Task struct {
	listenAddr string
	httpServer *http.Server
	taskTicker int64
	db         *db.WrapDb
}

func NewTask(cfg *config.ConfigApi, dao *db.WrapDb) (*Task, error) {
	s := &Task{
		listenAddr: cfg.ListenAddr,
		taskTicker: 10,
		db:         dao,
	}

	cache := map[string]uint64{
		utils.GateMinStakeAmount: cfg.GateMinStakeAmount,
	}

	handler := s.InitHandler(cache)

	s.httpServer = &http.Server{
		Addr:         s.listenAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return s, nil
}

func (svr *Task) InitHandler(cache map[string]uint64) http.Handler {
	return api.InitRouters(svr.db, cache)
}

func (svr *Task) ApiServer() {
	logrus.Infof("Gin server start on %s", svr.listenAddr)
	err := svr.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logrus.Errorf("Gin server start err: %s", err.Error())
		utils.ShutdownRequestChannel <- struct{}{} //shutdown server
		return
	}
	logrus.Infof("Gin server done on %s", svr.listenAddr)
}

func (svr *Task) Start() error {
	utils.SafeGoWithRestart(svr.ApiServer)
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
