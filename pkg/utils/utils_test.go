package utils_test

import (
	"encoding/json"
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
	questsBts, _ := json.Marshal(quests)
	t.Logf("quests: %s", string(questsBts))

	reviews, err := utils.GetCommunityReviews(apiKey, subdomain, userId)
	if err != nil {
		t.Fatal(err)
	}
	reviewsBts, _ := json.Marshal(reviews)
	t.Logf("reviews: %s", string(reviewsBts))

	for _, item := range reviews.Items {
		if item.Status == "success" {
			t.Log("success item", item.ID)
		}
	}

	user, err := utils.GetCommunityUser(apiKey, subdomain, ethAddress)
	if err != nil {
		t.Fatal(err)
	}
	userBts, _ := json.Marshal(user)
	t.Logf("user: %s", string(userBts))
}
