package utils_test

import (
	"invite-code-service/pkg/utils"
	"os"
	"testing"
)

func TestZealy(t *testing.T) {
	apiKey := os.Getenv("ZEALY_API_KEY")
	subdomain := os.Getenv("ZEALY_SUB_DOMAIN")
	userId := os.Getenv("USER_ID")
	ethAddress := os.Getenv("ETH_ADDRESS")
	quests, err := utils.GetCommunityQuests(apiKey, subdomain)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", quests)

	reviews, err := utils.GetCommunityReviews(apiKey, subdomain, userId)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", reviews)

	user, err := utils.GetCommunityUser(apiKey, subdomain, ethAddress)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", user)
}
