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
		Distances            []Distance `json:"distances"`
		Elevation            float64    `json:"elevation"`
		FairlyActiveMinutes  int64      `json:"fairlyActiveMinutes"`
		Floors               int64      `json:"floors"`
		LightlyActiveMinutes int64      `json:"lightlyActiveMinutes"`
		MarginalCalories     int64      `json:"marginalCalories"`
		SedentaryMinutes     int64      `json:"sedentaryMinutes"`
		Steps                int64      `json:"steps"`
		VeryActiveMinutes    int64      `json:"veryActiveMinutes"`
	}

	Distance struct {
		Activity string  `json:"activity"`
		Distance float64 `json:"distance"`
	}

	Activity struct {
		ActivityID       string  `json:"activityId"`
		ActivityParentID string  `json:"activityParentId"`
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
		Distance      float64 `json:"distance"`
		Floors        int64   `json:"floors"`
		Steps         int64   `json:"steps"`
	}

	ActivitiesResponse struct {
		Activities []Activity `json:"activities"`
		Goals      Goals      `json:"goals"`
		Summary    Summary    `json:"summary"`
	}

	User struct {
		AboutMe                string `json:"aboutMe"`
		Avatar150              string `json:"avatar150"`
		Avatar640              string `json:"avatar640"`
		City                   string `json:"city"`
		ClockTimeDisplayFormat string `json:"clockTimeDisplayFormat"`
		Country                string `json:"country"`
		DateOfBirth            string `json:"dateOfBirth"`
		DisplayName            string `json:"displayName"`
		DistanceUnit           string `json:"distanceUnit"`
		EncodedId              string `json:"encodedId"`
		FoodsLocale            string `json:"foodsLocale"`
		FullName               string `json:"fullName"`
		Gender                 string `json:"gender"`
		GlucoseUnitAboutMe     string `json:"glucoseUnit"`
		Height                 string `json:"height"`
		HeightUnit             string `json:"heightUnit"`
		MemberSince            string `json:"memberSince"`
		OffsetFromUTCMillis    string `json:"offsetFromUTCMillis"`
		StartDayOfWeek         string `json:"startDayOfWeek"`
		State                  string `json:"state"`
		StrideLengthRunning    string `json:"strideLengthRunning"`
		StrideLengthWalking    string `json:"strideLengthWalking"`
		Timezone               string `json:"timezone"`
		WaterUnit              string `json:"waterUnit"`
		Weight                 string `json:"weight"`
		WeightUnit             string `json:"weightUnit"`
	}

	UserResponse struct {
		User User `json:"user"`
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

func (s *API) User(userID, token string) (UserResponse, error) {
	var userReponse UserResponse

	resp, err := helpers.Get(s.Client, fmt.Sprintf("%s/user/%s/profile.json", s.URL, userID), token)
	if err != nil {
		return userReponse, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&userReponse)

	if resp.StatusCode != 200 {
		return userReponse, errors.New(resp.Status)
	}

	return userReponse, nil
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
