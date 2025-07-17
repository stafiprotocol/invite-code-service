package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"invite-code-service/dao"
	"invite-code-service/pkg/config"
	"invite-code-service/pkg/db"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func bindCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "bind",
		Short: "Bind code and address",

		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfigPath)
			if err != nil {
				return err
			}
			fmt.Printf("Config path: %s\n", configPath)

			cfg, err := config.LoadConfig[config.ConfigBindCode](configPath)
			if err != nil {
				return err
			}
			if len(cfg.FilePath) == 0 {
				return fmt.Errorf("FilePath empty")
			}

			file, err := os.Open(cfg.FilePath)
			if err != nil {
				return err
			}
			defer file.Close()

			reader := csv.NewReader(file)

			records, err := reader.ReadAll()
			if err != nil {
				return err
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

			for _, record := range records {
				if len(record) != 2 {
					return fmt.Errorf("format not match: recode: %s", record)
				}
				address := record[0]
				code := record[1]

				inviteCode, err := dao.GetInviteCode(db, code)
				if err != nil {
					if err != gorm.ErrRecordNotFound {
						return err
					}
					return fmt.Errorf("code %s not exist", code)
				} else {
					if inviteCode.UserAddress != nil {
						fmt.Printf("code %s already bind address: %s, will skip\n", code, address)
						continue
					}
				}
				if inviteCode.CodeType != dao.DirectInviteCode {
					return fmt.Errorf("code: %s type: %d not match", inviteCode.InviteCode, inviteCode.CodeType)
				}

				inviteCodeByUser, err := dao.GetInviteCodeByUserAddress(db, address)
				if err != nil {
					if err != gorm.ErrRecordNotFound {
						return err
					}
					// pass
				} else {
					fmt.Printf("address %s already bind code: %s, will skip\n", address, inviteCodeByUser.InviteCode)
					continue
				}

				inviteCode.UserAddress = &address
				inviteCode.BindTime = uint64(time.Now().Unix())
				err = dao.CheckBondAndUpdateInviteCode(db, inviteCode)
				if err != nil {
					return err
				}
				fmt.Printf("bind code: %s, address: %s, success\n", code, address)
			}

			return nil
		},
	}
	cmd.Flags().String(flagConfigPath, defaultConfigPath, "Config file path")
	return cmd
}
