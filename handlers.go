package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/angelacastanieto/hioqi/fitbitclient"
	"github.com/labstack/echo"
	"github.com/markbates/goth/gothic"
)

type GetUserResponse struct {
	StepsLeftToGo int64 `json:"steps_left_to_go"`
	CaloriesOut   int64 `json:"calories_out"`
	CaloriesGoal  int64 `json:"calories_goal"`
	StepsGoal     int64 `json:"steps_goal"`
	StepsSoFar    int64 `json:"steps_so_far"`
}

func GetUser(c echo.Context) error { // TODO:  this should be GetActivities
	id := c.Param("id")

	token, err := redisClient.Get(fmt.Sprintf("%s:access_token", id)).Result()
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	activitiesResponse, err := fitbitClient.Activities(id, time.Now().Format("2006-01-02"), token)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	getUserResponse := GetUserResponse{
		CaloriesGoal:  activitiesResponse.Goals.CaloriesOut,
		CaloriesOut:   activitiesResponse.Summary.CaloriesOut,
		StepsGoal:     activitiesResponse.Goals.Steps,
		StepsSoFar:    activitiesResponse.Summary.Steps,
		StepsLeftToGo: stepsLeftToGo(activitiesResponse.Goals.CaloriesOut, activitiesResponse.Summary.CaloriesOut, caloriesOutPerStep(activitiesResponse.Activities, activitiesResponse.Summary.CaloriesOut, activitiesResponse.Summary.Steps)),
	}

	return c.JSON(http.StatusOK, getUserResponse)
}

func AuthHandler(c echo.Context) error {
	gothic.BeginAuthHandler(c.Response(), c.Request())

	return nil
}

func CallbackHandler(c echo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		fmt.Println(err)
		return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000"+user.UserID)
	}

	err = redisClient.Set(fmt.Sprintf("%s:access_token", user.UserID), user.AccessToken, -time.Since(user.ExpiresAt)).Err()
	if err != nil {
		fmt.Println(err)
		return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000"+user.UserID)
	}

	err = redisClient.Set(fmt.Sprintf("%s:refresh_token", user.UserID), user.RefreshToken, -time.Since(user.ExpiresAt)).Err()
	if err != nil {
		fmt.Println(err)
		return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000"+user.UserID)
	}

	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000/users/"+user.UserID)
}

func caloriesOutPerStep(activities []fitbitclient.Activity, caloriesOut, stepsSoFar int64) float64 {
	// need to subtract BMR calories
	// also need to get this calc from last day on which you took steps
	var stepsFromActivities int64
	var caloriesOutFromActivities int64
	for _, activity := range activities {
		stepsFromActivities += activity.Steps
		caloriesOutFromActivities += activity.Calories
	}

	stepsNonActivities := stepsSoFar - stepsFromActivities
	caloriesOutNonActivities := caloriesOut - caloriesOutFromActivities

	return float64(caloriesOutNonActivities / stepsNonActivities)
}

func stepsLeftToGo(caloriesOutGoal, caloriesOut int64, caloriesOutPerStep float64) int64 {
	caloriesLeftToBurn := caloriesOutGoal - caloriesOut
	return int64(float64(caloriesLeftToBurn) / caloriesOutPerStep)
}
