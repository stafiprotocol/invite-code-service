package api

import (
	"fmt"
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

const maxGenCount = 100000

type Service struct {
	cfg *config.ConfigApi

	httpServer *http.Server
	db         *db.WrapDb
}

func NewService(cfg *config.ConfigApi, dao *db.WrapDb) (*Service, error) {
	if cfg.TaskInviteCodeCount > maxGenCount || cfg.DirectInviteCodeCount > maxGenCount {
		return nil, fmt.Errorf("over max gen count: %d", maxGenCount)
	}

	s := &Service{
		cfg: cfg,
		db:  dao,
	}

	handler := s.InitHandler()

	s.httpServer = &http.Server{
		Addr:         s.cfg.ListenAddr,
		Handler:      handler,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	return s, nil
}

func (svr *Service) InitHandler() http.Handler {
	return api.InitRouters(svr.db, svr.cfg)
}

func (svr *Service) ApiServer() {
	logrus.Infof("Gin server start on %s", svr.cfg.ListenAddr)
	err := svr.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logrus.Errorf("Gin server start err: %s", err.Error())
		utils.ShutdownRequestChannel <- struct{}{} //shutdown server
		return
	}
	logrus.Infof("Gin server done on %s", svr.cfg.ListenAddr)
}

func (svr *Service) Start() error {
	taskInviteCodeCount, err := dao.GetInviteCodeCount(svr.db, dao.TaskInviteCode)
	if err != nil {
		return err
	}
	directInviteCodeCount, err := dao.GetInviteCodeCount(svr.db, dao.DirectInviteCode)
	if err != nil {
		return err
	}
	waterInviteCodeCount, err := dao.GetInviteCodeCount(svr.db, dao.WaterInviteCode)
	if err != nil {
		return err
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

	if directInviteCodeCount < int64(svr.cfg.DirectInviteCodeCount) {
		genCount := int64(svr.cfg.DirectInviteCodeCount) - directInviteCodeCount
		logrus.Infof("need generate %d direct invite code", genCount)
		err := svr.genInviteCode(genCount, dao.DirectInviteCode)
		if err != nil {
			return err
		}
		logrus.Infof("generate success")
	}

	if svr.cfg.DropletRound > 0 {
		var maxRound uint8
		err = svr.db.Model(&dao.DropletCode{}).
			Select("MAX(round)").
			Scan(&maxRound).Error
		if err != nil {
			return fmt.Errorf("failed to get max round: %w", err)
		}

		if svr.cfg.DropletRound > maxRound+1 {
			return fmt.Errorf("exist max round: %d", maxRound)
		}
	}

	needWaterInviteCodeCount := uint64(svr.cfg.DropletRound+1) * utils.DropletCount * utils.CodesPerDroplet

	if waterInviteCodeCount < int64(needWaterInviteCodeCount) {
		genCount := int64(needWaterInviteCodeCount) - waterInviteCodeCount

		logrus.Infof("need generate %d water invite code", genCount)
		err := svr.genInviteCode(genCount, dao.WaterInviteCode)
		if err != nil {
			return err
		}
		logrus.Infof("generate success")
	}

	err = dao.GenerateDropletCodes(svr.db, svr.cfg.DropletRound)
	if err != nil {
		return fmt.Errorf("GenerateDropletCodes failed: %s", err.Error())
	}

	utils.SafeGoWithRestart(svr.ApiServer)
	return nil
}

func (svr *Service) genInviteCode(genCount int64, codeType uint8) error {
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

		err = dao.CreateInviteCode(svr.db, &newInviteCode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svr *Service) Stop() {
	if svr.httpServer != nil {
		err := svr.httpServer.Close()
		if err != nil {
			logrus.Errorf("Problem shutdown Gin server :%s", err.Error())
		}
	}
}
