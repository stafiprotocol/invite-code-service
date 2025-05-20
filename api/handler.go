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
	cacheKeyUserInfo = "cacheKeyUserInfo_%s"
	cacheKeyUserTask = "cacheKeyUserTask_%s"
)

func userInfoKey(addr string) string {
	return fmt.Sprintf(cacheKeyUserInfo, addr)
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

func (h *Handler) getUserInfo(address string) (*utils.UserResponse, error) {
	cachedUserInfo, found := h.cache.Get(userInfoKey(address))
	if !found {
		user, err := utils.GetCommunityUser(h.cfg.ZealyApiKey, h.cfg.ZealySubdomain, address)
		if err != nil {
			return nil, err
		}

		if len(user.DiscordID) > 0 {
			cachedUserInfo = user
			h.cache.Set(userInfoKey(address), user, cache.NoExpiration)
		}

	}

	user, ok := cachedUserInfo.(*utils.UserResponse)
	if !ok {
		return nil, fmt.Errorf("cast cachedUserId failed, %+v", cachedUserInfo)
	}

	if len(user.ID) == 0 {
		return nil, fmt.Errorf("user id empty")
	}

	return user, nil
}

func (h *Handler) getUserTasks(address string) ([]Task, error) {
	cachedUserTask, found := h.cache.Get(userTaskKey(address))
	if !found {
		userInfo, err := h.getUserInfo(address)
		if err != nil {
			return nil, err
		}

		reviews, err := utils.GetCommunityReviews(h.cfg.ZealyApiKey, h.cfg.ZealySubdomain, userInfo.ID)
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
		if item.Status == "success" {
			tasks = append(tasks, Task{
				Id:          item.Quest.ID,
				Description: item.Quest.Name,
			})
		}
	}

	return tasks, nil
}
