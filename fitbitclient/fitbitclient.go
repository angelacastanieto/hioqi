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

	Summary struct {
		ActivityCalories     int64      `json:"activityCalories"`
		CaloriesBMR          int64      `json:"caloriesBMR"`
		CaloriesOut          int64      `json:"caloriesOut"`
		distances            []Distance `json:"distances"`
		elevation            float64    `json:"elevation"`
		fairlyActiveMinutes  int64      `json:"fairlyActiveMinutes"`
		floors               int64      `json:"floors"`
		lightlyActiveMinutes int64      `json:"lightlyActiveMinutes"`
		marginalCalories     int64      `json:"marginalCalories"`
		sedentaryMinutes     int64      `json:"sedentaryMinutes"`
		steps                int64      `json:"steps"`
		veryActiveMinutes    int64      `json:"veryActiveMinutes"`
	}

	Distance struct {
		Activity string  `json:"activity"`
		Distance float64 `json:"distance"`
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
		StartTime        string  `json:"startTime"`
		Steps            int64   `json:"steps"`
	}

	Goals struct {
		ActiveMinutes int64   `json:"activeMinutes"`
		CaloriesOut   int64   `json:"caloriesOut"`
		distance      float64 `json:"distance"`
		floors        int64   `json:"floors"`
		steps         int64   `json:"steps"`
	}

	ActivitiesResponse struct {
		Activities []Activity `json:"activities"`
		Goals      Goals      `json:"goals"`
		Summary    Summary    `json:"summary"`
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
