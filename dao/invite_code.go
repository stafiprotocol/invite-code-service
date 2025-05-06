package dao

import (
	"invite-code-service/pkg/db"
)

const (
	TaskInviteCode   = uint8(0)
	DirectInviteCode = uint8(1)
)

type InviteCode struct {
	db.BaseModel

	InviteCode  string  `gorm:"type:varchar(10);not null;default:'';column:invite_code;uniqueIndex"`
	UserAddress *string `gorm:"type:varchar(80);column:user_address;uniqueIndex"`

	CodeType uint8  `gorm:"type:tinyint(1);unsigned;not null;default:0;column:code_type"`
	BindTime uint64 `gorm:"type:int(11);unsigned;not null;default:0;column:bind_time"`
}

func (f InviteCode) TableName() string {
	return "invite_codes"
}

func UpOrInInviteCode(db *db.WrapDb, c *InviteCode) error {
	return db.Save(c).Error
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

func GetInviteCodeByUser(db *db.WrapDb, user string) (info *InviteCode, err error) {
	info = &InviteCode{}
	err = db.Take(info, "user_address = ?", user).Error
	return
}

func GetAvailableTaskInviteCode(db *db.WrapDb) (info *InviteCode, err error) {
	info = &InviteCode{}
	err = db.Take(info, "code_type = 0 && bind_time = 0").Error
	return
}
