package dao

import (
	"fmt"
	"invite-code-service/pkg/db"
	"invite-code-service/pkg/utils"

	"gorm.io/gorm"
)

type DropletCode struct {
	db.BaseModel

	InviteCode   string `gorm:"type:varchar(10);not null;default:'';column:invite_code;uniqueIndex:code_round_index"`
	Round        uint8  `gorm:"type:tinyint(1);unsigned;not null;default:0;column:round;uniqueIndex:code_round_index"`
	DropletIndex uint8  `gorm:"type:tinyint(1);unsigned;not null;default:0;column:droplet_index;uniqueIndex:code_round_index"`
}

func (f DropletCode) TableName() string {
	return "droplet_codes"
}

type DropletCodeWithStatus struct {
	InviteCode   string
	Round        uint8
	DropletIndex uint8
	Used         bool
}

func GetLatestDropletCodesWithStatus(db *db.WrapDb) ([]*DropletCodeWithStatus, error) {
	var maxRound uint8
	err := db.Model(&DropletCode{}).
		Select("MAX(round)").
		Scan(&maxRound).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get max round: %w", err)
	}

	var dropletCodes []DropletCode
	err = db.Where("round = ?", maxRound).
		Find(&dropletCodes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet codes: %w", err)
	}

	if len(dropletCodes) == 0 {
		return nil, nil
	}

	inviteCodes := make([]string, 0, len(dropletCodes))
	for _, dc := range dropletCodes {
		inviteCodes = append(inviteCodes, dc.InviteCode)
	}

	var usedCodes []string
	err = db.Model(&InviteCode{}).
		Where("invite_code IN ?", inviteCodes).
		Where("bind_time > 0").
		Pluck("invite_code", &usedCodes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query invite code usage: %w", err)
	}

	usedSet := make(map[string]struct{}, len(usedCodes))
	for _, code := range usedCodes {
		usedSet[code] = struct{}{}
	}

	var result []*DropletCodeWithStatus
	for _, dc := range dropletCodes {
		_, used := usedSet[dc.InviteCode]
		result = append(result, &DropletCodeWithStatus{
			InviteCode:   dc.InviteCode,
			Round:        dc.Round,
			DropletIndex: dc.DropletIndex,
			Used:         used,
		})
	}

	return result, nil
}

func GenerateDropletCodes(db *db.WrapDb, round uint8) error {
	var dropletCodes []DropletCode
	err := db.Where("round = ?", round).
		Find(&dropletCodes).Error
	if err != nil {
		return fmt.Errorf("failed to get droplet codes: %w", err)
	}
	if len(dropletCodes) > 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {

		// Fetch enough InviteCodes
		totalNeeded := utils.DropletCount * utils.CodesPerDroplet
		var availableCodes []InviteCode
		if err := tx.
			Where("code_type = 2 AND bind_time = 0").
			Order("id ASC").
			Limit(totalNeeded).
			Find(&availableCodes).Error; err != nil {
			return fmt.Errorf("failed to fetch droplet invite codes: %w", err)
		}

		if len(availableCodes) < totalNeeded {
			return fmt.Errorf("not enough available droplet invite codes")
		}

		// Assign codes and create droplets
		cursor := 0
		for dropletIdx := uint8(0); dropletIdx < utils.DropletCount; dropletIdx++ {
			for i := 0; i < utils.CodesPerDroplet; i++ {
				code := availableCodes[cursor]
				cursor++

				droplet := DropletCode{
					InviteCode:   code.InviteCode,
					Round:        round,
					DropletIndex: dropletIdx,
				}
				if err := tx.Create(&droplet).Error; err != nil {
					return fmt.Errorf("failed to create droplet: %w", err)
				}
			}
		}

		return nil
	})
}
