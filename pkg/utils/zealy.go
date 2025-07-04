package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type QuestResponse []struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Description       interface{}   `json:"description"`
	CommunityID       string        `json:"communityId"`
	CategoryID        string        `json:"categoryId"`
	Deleted           bool          `json:"deleted"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
	Archived          bool          `json:"archived"`
	AutoValidate      bool          `json:"autoValidate"`
	Conditions        []interface{} `json:"conditions"`
	ConditionOperator string        `json:"conditionOperator"`
	Published         bool          `json:"published"`
	Recurrence        string        `json:"recurrence"`
	RetryAfter        int           `json:"retryAfter"`
	Rewards           []struct {
		Type   string `json:"type"`
		Value  int    `json:"value"`
		Method struct {
			Type string `json:"type"`
		} `json:"method"`
	} `json:"rewards"`
	Tasks             []TaskDetail `json:"tasks"`
	ClaimCounter      int          `json:"claimCounter"`
	CommunityImageURL interface{}  `json:"communityImageUrl"`
	CommunityName     string       `json:"communityName"`
	Subdomain         string       `json:"subdomain"`
}

type TaskDetail struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Metadata struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		GuildID  string `json:"guildId"`
		ImageURL string `json:"imageUrl"`
	} `json:"metadata"`
	Settings struct {
		InviteURL string `json:"inviteUrl"`
		Username  string `json:"username"`
	} `json:"settings"`
}

func GetCommunityQuests(apiKey, subdomain string) (QuestResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api-v2.zealy.io/public/communities/%s/quests", subdomain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var quests QuestResponse
	err = json.Unmarshal(body, &quests)
	if err != nil {
		return nil, err
	}

	return quests, nil
}

type ReviewResponse struct {
	Items      []ReviewItem `json:"items"`
	NextCursor string       `json:"nextCursor"`
}

type ReviewItem struct {
	ID             string `json:"id"`
	User           User   `json:"user"`
	Quest          Quest  `json:"quest"`
	Status         string `json:"status"`
	Mark           string `json:"mark"`
	LastReviewerID string `json:"lastReviewerId"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
	Tasks          []Task `json:"tasks"`
	AutoValidated  bool   `json:"autoValidated"`
}

type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type Quest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Task struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	Status    string `json:"status"`
	Type      string `json:"type"`
}

func GetCommunityReviews(apiKey, subdomain, userId string) (*ReviewResponse, error) {
	url := fmt.Sprintf("https://api-v2.zealy.io/public/communities/%s/reviews", subdomain)
	if userId != "" {
		url += "?userId=" + userId
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var response ReviewResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type UserResponse struct {
	DiscordHandle                 string            `json:"discordHandle"`
	TiktokUsername                string            `json:"tiktokUsername"`
	TwitterUsername               string            `json:"twitterUsername"`
	VerifiedBlockchainAddresses   map[string]string `json:"verifiedBlockchainAddresses"`
	UnVerifiedBlockchainAddresses map[string]string `json:"unVerifiedBlockchainAddresses"`
	ConnectedWallet               string            `json:"connectedWallet"`
	Email                         string            `json:"email"`
	DiscordID                     string            `json:"discordId"`
	TwitterID                     string            `json:"twitterId"`
	ID                            string            `json:"id"`
	XP                            int               `json:"xp"`
	Name                          string            `json:"name"`
	CreatedAt                     string            `json:"createdAt"`
	Rank                          int               `json:"rank"`
	Invites                       []Invite          `json:"invites"`
	Role                          string            `json:"role"`
	Level                         int               `json:"level"`
	IsBanned                      bool              `json:"isBanned"`
	Karma                         int               `json:"karma"`
	ReferrerURL                   string            `json:"referrerUrl"`
	ReferrerID                    string            `json:"referrerId"`
	BanReason                     string            `json:"banReason"`
}

type Invite struct {
	UserID   string `json:"userId"`
	Status   string `json:"status"`
	JoinedAt string `json:"joinedAt"`
	XP       int    `json:"xp"`
}

var ErrAddressNotFound = errors.New("ErrAddressNotFound")

func GetCommunityUser(apiKey, subdomain, ethAddress string) (*UserResponse, error) {
	url := fmt.Sprintf("https://api-v2.zealy.io/public/communities/%s/users", subdomain)
	if ethAddress != "" {
		url += "?ethAddress=" + ethAddress
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrAddressNotFound
		}

		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
