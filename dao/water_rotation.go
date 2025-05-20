package dao

import (
	"errors"
	"invite-code-service/pkg/db"
	"time"

	"gorm.io/gorm"
)

type WaterRotation struct {
	db.BaseModel

	InviteCode string `gorm:"type:varchar(10);not null;default:'';column:invite_code;uniqueIndex"`
}

func (f WaterRotation) TableName() string {
	return "water_rotations"
}

func UpOrInWaterRotation(db *db.WrapDb, c *WaterRotation) error {
	return db.Save(c).Error
}

func GetWaterRotations(db *db.WrapDb) (rotations []*WaterRotation, err error) {
	err = db.Find(&rotations).Error
	return
}

const waterRotationsCountLimit = 10
const waterRotationsRefreshSeconds = 60 * 60

func TryRotateInviteCodes(db *db.WrapDb) error {
	var state WaterRotation
	err := db.First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return RefreshInviteCodeRotation(db)
	} else if err != nil {
		return err
	}

	if time.Now().Unix() >= int64(waterRotationsRefreshSeconds+state.CreatedAt) {
		if err := RefreshInviteCodeRotation(db); err != nil {
			return err
		}
	}
	return nil
}

func RefreshInviteCodeRotation(db *db.WrapDb) error {
	if err := db.Exec("DELETE FROM water_rotations").Error; err != nil {
		return err
	}

	var codes []InviteCode
	err := db.Where("bind_time = 0 AND code_type = 2").
		Order("RAND()").Limit(waterRotationsCountLimit).
		Find(&codes).Error
	if err != nil {
		return err
	}

	var rotations []WaterRotation
	for _, c := range codes {
		rotations = append(rotations, WaterRotation{
			InviteCode: c.InviteCode,
		})
	}
	return db.Create(&rotations).Error
}

type WaterRotationWithStatus struct {
	InviteCode string
	Used       bool
}

func GetWaterRotationsWithStatus(db *db.WrapDb) (result []WaterRotationWithStatus, err error) {
	err = db.Model(WaterRotation{}).
		Select("water_rotations.invite_code, IF(invite_codes.bind_time > 0, true, false) AS used").
		Joins("LEFT JOIN invite_codes ON water_rotations.invite_code = invite_codes.invite_code").
		Scan(&result).Error

	return
}
