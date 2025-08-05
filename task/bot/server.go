package bot

import (
	"fmt"
	"invite-code-service/dao"
	"invite-code-service/pkg/config"
	"invite-code-service/pkg/db"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Task struct {
	cfg *config.ConfigDiscordBot

	db *db.WrapDb

	discordClient *discordgo.Session
}

func NewTask(cfg *config.ConfigDiscordBot, dao *db.WrapDb) (*Task, error) {
	dg, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		return nil, err
	}

	return &Task{cfg: cfg, db: dao, discordClient: dg}, nil
}

func (svr *Task) Start() error {
	svr.discordClient.AddHandler(svr.claimRoleHandler)

	svr.discordClient.Identify.Intents = discordgo.IntentsGuildMessages

	return svr.discordClient.Open()
}

func (svr *Task) Stop() {
	svr.discordClient.Close()
}

func (svr *Task) claimRoleHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.ChannelID != svr.cfg.DiscordChannelId {
		return
	}

	content := m.Content
	logrus.Infof("msg received, content: %s", content)

	re := regexp.MustCompile(`\s+`)
	subs := re.Split(content, -1)

	if len(subs) != 2 || subs[0] != "!claim" || subs[1] != "og" {
		logrus.Warnf("Illegal format: %s", content)
		return
	}

	_, err := dao.GetInviteCodeByDiscordId(svr.db, m.Author.ID)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			logrus.Errorf("GetInviteCodeByDiscordId error: %s", err.Error())
			return
		}

		reMsg := fmt.Sprintf("%s(%s) failed to claim (invite code not found)", m.Author.DisplayName(), m.Author.Username)
		_, err = s.ChannelMessageSend(m.ChannelID, reMsg)
		if err != nil {
			logrus.Errorf("discordBot send msg error: %s", err.Error())
			return
		}
		logrus.Warnf("user: %s has no code", m.Author.ID)
		return
	}

	err = svr.discordClient.GuildMemberRoleAdd(svr.cfg.DiscordGuidId, m.Author.ID, svr.cfg.DiscordRoleId)
	if err != nil {
		logrus.Errorf("GuildMemberRoleAdd error: %s", err.Error())
		return
	}

	reMsg := fmt.Sprintf("%s(%s) claimed success", m.Author.DisplayName(), m.Author.Username)
	_, err = s.ChannelMessageSend(m.ChannelID, reMsg)
	if err != nil {
		logrus.Errorf("discordBot send msg error: %s", err.Error())
		return
	}
}
