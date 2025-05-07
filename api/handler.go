package api

import (
	"fmt"
	"invite-code-service/pkg/config"
	"invite-code-service/pkg/db"
	"invite-code-service/pkg/utils"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	codeParamErr                  = "80001"
	codeInternalErr               = "80002"
	codeUserAlreadyBoundErr       = "80003"
	codeInviteCodeAlreadyBoundErr = "80004"
	codeUserSigVerifyErr          = "80005"
	codeUserTaskVerifyErr         = "80006"
	codeInviteCodeNotExistErr     = "80007"
	codeInviteCodeTypeNotMatchErr = "80008"
	codeInviteCodeNotEnoughErr    = "80009"
)

const (
	cacheKeyTask     = "cacheKeyTask"
	cacheKeyUserId   = "cacheKeyUserId_%s"
	cacheKeyUserTask = "cacheKeyUserTask_%s"
)

func userIdKey(addr string) string {
	return fmt.Sprintf(cacheKeyUserId, addr)
}
func userTaskKey(addr string) string {
	return fmt.Sprintf(cacheKeyUserTask, addr)
}

type Handler struct {
	db    *db.WrapDb
	cfg   *config.ConfigApi
	cache *cache.Cache
}

func NewHandler(db *db.WrapDb, cfg *config.ConfigApi) *Handler {
	return &Handler{db: db, cfg: cfg, cache: cache.New(time.Minute*10, time.Minute*1)}
}

func (h *Handler) getTasks() ([]Task, error) {
	cachedTask, found := h.cache.Get(cacheKeyTask)
	if !found {
		res, err := utils.GetCommunityQuests(h.cfg.ZealyApiKey, h.cfg.ZealySubdomain)
		if err != nil {
			return nil, err

		}

		cachedTask = res

		h.cache.Set(cacheKeyTask, res, cache.DefaultExpiration)
	}

	quests, ok := cachedTask.(utils.QuestResponse)
	if !ok {
		return nil, fmt.Errorf("cast cachedTask failed, %+v", cachedTask)
	}

	tasks := make([]Task, 0, len(quests))
	for _, quest := range quests {
		if quest.Published {
			tasks = append(tasks, Task{
				Id:          quest.ID,
				Description: quest.Name,
			})
		}
	}

	return tasks, nil
}

func (h *Handler) getUserId(address string) (*string, error) {
	cachedUserId, found := h.cache.Get(userIdKey(address))
	if !found {
		user, err := utils.GetCommunityUser(h.cfg.ZealyApiKey, h.cfg.ZealySubdomain, address)
		if err != nil {
			return nil, err
		}

		cachedUserId = user.ID

		h.cache.Set(userIdKey(address), user.ID, cache.NoExpiration)
	}

	userId, ok := cachedUserId.(string)
	if !ok {
		return nil, fmt.Errorf("cast cachedUserId failed, %+v", cachedUserId)
	}

	return &userId, nil
}

func (h *Handler) getUserTasks(address string) ([]Task, error) {
	cachedUserTask, found := h.cache.Get(userTaskKey(address))
	if !found {
		userId, err := h.getUserId(address)
		if err != nil {
			return nil, err
		}

		reviews, err := utils.GetCommunityReviews(h.cfg.ZealyApiKey, h.cfg.ZealySubdomain, *userId)
		if err != nil {
			return nil, err
		}

		cachedUserTask = reviews

		h.cache.Set(userTaskKey(address), reviews, time.Second)
	}

	userTask, ok := cachedUserTask.(*utils.ReviewResponse)
	if !ok {
		return nil, fmt.Errorf("cast cachedUserTask failed, %+v", cachedUserTask)
	}

	tasks := make([]Task, 0, len(userTask.Items))
	for _, item := range userTask.Items {
		tasks = append(tasks, Task{
			Id:          item.Quest.ID,
			Description: item.Quest.Name,
		})
	}

	return tasks, nil
}
