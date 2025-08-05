package cmd

import (
	"encoding/json"
	"fmt"
	"invite-code-service/dao"
	"invite-code-service/pkg/config"
	"invite-code-service/pkg/db"
	"invite-code-service/pkg/log"
	"invite-code-service/pkg/utils"
	task "invite-code-service/task/bot"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func startDiscordBotCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start-discord-bot",
		Short: "Start discord bot",

		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfigPath)
			if err != nil {
				return err
			}
			fmt.Printf("Config path: %s\n", configPath)

			cfg, err := config.LoadConfig[config.ConfigDiscordBot](configPath)
			if err != nil {
				return err
			}
			if len(cfg.LogFileDir) == 0 {
				cfg.LogFileDir = "./log_data"
			}

			bts, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Printf("Config: \n%s\n", string(bts))
		Out:
			for {
				fmt.Println("\nCheck config info, then press (y/n) to continue:")
				var input string
				fmt.Scanln(&input)
				switch input {
				case "y":
					break Out
				case "n":
					return nil
				default:
					fmt.Println("press `y` or `n`")
					continue
				}
			}

			logLevelStr, err := cmd.Flags().GetString(flagLogLevel)
			if err != nil {
				return err
			}
			logLevel, err := logrus.ParseLevel(logLevelStr)
			if err != nil {
				return err
			}

			logrus.SetLevel(logLevel)
			err = log.InitLogFile(cfg.LogFileDir + "/bot")
			if err != nil {
				return fmt.Errorf("InitLogFile failed: %w", err)
			}

			//init db
			db, err := db.NewDB(&db.Config{
				Host:   cfg.Db.Host,
				Port:   cfg.Db.Port,
				User:   cfg.Db.User,
				Pass:   cfg.Db.Pwd,
				DBName: cfg.Db.Name,
				Mode:   "info"})
			if err != nil {
				logrus.Errorf("db err: %s", err)
				return err
			}
			err = dao.AutoMigrate(db)
			if err != nil {
				logrus.Errorf("dao autoMigrate err: %s", err)
				return err
			}
			logrus.Infof("db connect success")

			ctx := utils.ShutdownListener()

			t, err := task.NewTask(cfg, db)
			if err != nil {
				return err
			}
			err = t.Start()
			if err != nil {
				return err
			}

			defer func() {
				logrus.Infof("shutting down task ...")
				t.Stop()
			}()

			<-ctx.Done()

			return nil
		},
	}
	cmd.Flags().String(flagConfigPath, defaultConfigPath, "Config file path")
	cmd.Flags().String(flagLogLevel, logrus.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
	return cmd
}
