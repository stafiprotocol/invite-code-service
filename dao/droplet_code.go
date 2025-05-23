package dao

import (
	"fmt"
	"invite-code-service/pkg/db"
	"time"

	"gorm.io/gorm"
)

const roundRefreshSeconds = 7 * 24 * 60 * 60 // 1 week
const (
	dropletCount    = 5
	codesPerDroplet = 5
	totalPerRound   = dropletCount * codesPerDroplet
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

func GenerateDropletCodes(db *db.WrapDb) error {

	now := time.Now()

	return db.Transaction(func(tx *gorm.DB) error {
		var existingDropletCodes []DropletCode
		if err := tx.Order("round ASC, droplet_index ASC").Find(&existingDropletCodes).Error; err != nil {
			return fmt.Errorf("failed to query existing droplets: %w", err)
		}

		// Build map[round][group]
		dropletCodeMap := make(map[uint8]map[uint8][]DropletCode)
		for _, d := range existingDropletCodes {
			if dropletCodeMap[d.Round] == nil {
				dropletCodeMap[d.Round] = make(map[uint8][]DropletCode)
			}
			dropletCodeMap[d.Round][d.DropletIndex] = append(dropletCodeMap[d.Round][d.DropletIndex], d)
		}

		var round uint8
		var dropletsToCreate []uint8

		if len(dropletCodeMap[0]) < dropletCount {
			// First time: round 0, initialize full groups
			round = 0
			for i := 0; i < dropletCount; i++ {
				if _, exists := dropletCodeMap[0][uint8(i)]; !exists {
					dropletsToCreate = append(dropletsToCreate, uint8(i))
				}
			}
		} else {
			// round = 1, must check one week delay
			round = 1

			for i := 0; i < dropletCount; i++ {
				idx := uint8(i)

				codesOfDropletRound0 := dropletCodeMap[0][idx]
				codesOfDropletRound1 := dropletCodeMap[1][idx]

				// Must exist in round 0
				if len(codesOfDropletRound0) != codesPerDroplet {
					continue
				}

				// Already created in round 1
				if len(codesOfDropletRound1) == codesPerDroplet {
					continue
				}

				// Check if all round 0 droplets older than 1 week
				if now.Unix() > int64(codesOfDropletRound0[0].CreatedAt+roundRefreshSeconds) {
					dropletsToCreate = append(dropletsToCreate, idx)
				}
			}
		}

		if len(dropletsToCreate) == 0 {
			return nil
		}

		// Fetch enough InviteCodes
		totalNeeded := len(dropletsToCreate) * codesPerDroplet
		var availableCodes []InviteCode
		if err := tx.
			Where("code_type = 2 AND bind_time = 0").
			Order("id ASC").
			Limit(totalNeeded).
			Find(&availableCodes).Error; err != nil {
			return fmt.Errorf("failed to fetch invite codes: %w", err)
		}

		if len(availableCodes) < totalNeeded {
			return fmt.Errorf("not enough available invite codes")
		}

		// Assign codes and create droplets
		cursor := 0
		for _, dropletIdx := range dropletsToCreate {
			for i := 0; i < codesPerDroplet; i++ {
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
