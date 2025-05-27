package dao

import (
	"errors"
	"invite-code-service/pkg/db"
)

const (
	TaskInviteCode   = uint8(0)
	DirectInviteCode = uint8(1)
	WaterInviteCode  = uint8(2)
)

type InviteCode struct {
	db.BaseModel

	InviteCode string `gorm:"type:varchar(10);not null;default:'';column:invite_code;uniqueIndex"`

	UserAddress *string `gorm:"type:varchar(80);column:user_address;uniqueIndex"`
	DiscordId   *string `gorm:"type:varchar(80);column:discord_id;uniqueIndex"`
	DiscordName *string `gorm:"type:varchar(80);column:discord_name;"`
	UserId      *string `gorm:"type:varchar(80);column:user_id;uniqueIndex"`

	CodeType uint8  `gorm:"type:tinyint(1);unsigned;not null;default:0;column:code_type"`
	BindTime uint64 `gorm:"type:int(11);unsigned;not null;default:0;column:bind_time"`
}

func (f InviteCode) TableName() string {
	return "invite_codes"
}

func CreateInviteCode(db *db.WrapDb, c *InviteCode) error {
	return db.Save(c).Error
}

var ErrAlreadyBond = errors.New("already bond")

func CheckBondAndUpdateInviteCode(db *db.WrapDb, c *InviteCode) error {
	result := db.Model(c).Where("bind_time = 0").Select("*").Omit("CreatedAt", "InviteCode", "CodeType").Updates(c)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAlreadyBond
	}
	return nil
}

func GetInviteCode(db *db.WrapDb, code string) (info *InviteCode, err error) {
	info = &InviteCode{}
	err = db.Take(info, "invite_code = ?", code).Error
	return
}

func GetInviteCodeCount(db *db.WrapDb, codeType uint8) (count int64, err error) {
	err = db.Model(&InviteCode{}).Where("code_type = ?", codeType).Count(&count).Error
	return
}

func GetInviteCodeByUserAddress(db *db.WrapDb, user string) (info *InviteCode, err error) {
	info = &InviteCode{}
	err = db.Take(info, "user_address = ?", user).Error
	return
}

func GetInviteCodeByDiscordId(db *db.WrapDb, discordId string) (info *InviteCode, err error) {
	info = &InviteCode{}
	err = db.Take(info, "discord_id = ?", discordId).Error
	return
}

func GetAvailableTaskInviteCode(db *db.WrapDb) (info *InviteCode, err error) {
	info = &InviteCode{}
	err = db.Where("code_type = 0 AND bind_time = 0").Order("RAND()").First(info).Error
	return
}

type InviteCodeStats struct {
	TotalCodes  int64 `json:"totalCodes"`
	RemainCodes int64 `json:"remainCodes"`
}

func GetInviteCodeStats(db *db.WrapDb) (*InviteCodeStats, error) {
	var total int64
	var unused int64

	if err := db.Model(&InviteCode{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&InviteCode{}).Where("bind_time = 0").Count(&unused).Error; err != nil {
		return nil, err
	}

	return &InviteCodeStats{
		TotalCodes:  total,
		RemainCodes: unused,
	}, nil
}
