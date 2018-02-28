package fitbitclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/angelacastanieto/hioqi/helpers"
)

type (
	API struct {
		Client *http.Client
		URL    string
	}

	Activity struct {
		ActivityID       string  `json:"activityId"`
		CctivityParentID string  `json:"activityParentId"`
		Calories         int64   `json:"calories"`
		Distance         float64 `json:"distance"`
		Duration         int64   `json:"duration"`
		IsFavorite       bool    `json:"isFavorite"`
		LogID            string  `json:"logId"`
		Name             string  `json:"name"`

		StartTime string `json:"startTime"`

		Steps int64 `json:"steps"`
	}

	ActivitiesResponse struct {
		Activities []Activity `json:"activities"`
	}
)

const (
	URL = "https://api.fitbit.com/1"
)

func NewAPI() (*API, error) {
	return &API{
		Client: &http.Client{},
		URL:    URL,
	}, nil
}

func (s *API) Activities(userID, dateString, token string) (ActivitiesResponse, error) {
	var activitiesResponse ActivitiesResponse

	resp, err := helpers.Get(s.Client, fmt.Sprintf("%s/user/%s/activities/date/%s.json", s.URL, userID, dateString), token)
	if err != nil {
		return activitiesResponse, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&activitiesResponse)

	if resp.StatusCode != 200 {
		return activitiesResponse, errors.New(resp.Status)
	}

	return activitiesResponse, nil
}
