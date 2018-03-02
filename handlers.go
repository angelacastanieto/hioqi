package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/markbates/goth/gothic"
)

const (
	IntensityMaintenance = "MAINTENANCE"
	IntensityEasier      = "EASIER"
	IntensityMedium      = "MEDIUM"
	IntensityKindaHard   = "KINDAHARD"
	IntensityHarder      = "HARDER"

	IntensityMaintenanceCalories = 0
	IntensityEasierCalories      = 250
	IntensityMediumCalories      = 500
	IntensityKindaHardCalories   = 750
	IntensityHarderCalories      = 1000
)

type GetUserResponse struct {
	Avatar             string `json:"avatar"`
	Name               string `json:"name"`
	StepsLeftToGo      int64  `json:"steps_left_to_go"`
	CaloriesLeftToGo   int64  `json:"calories_left_to_go"`
	CalorieDeficitGoal int64  `json:"calorie_deficit_goal"`
	CaloriesGoal       int64  `json:"calories_goal"`
	CaloriesIn         int64  `json:"calories_in"`
	CaloriesOut        int64  `json:"calories_out"`
	StepsGoal          int64  `json:"steps_goal"`
	StepsSoFar         int64  `json:"steps_so_far"`
}

var (
	ErrInvalidIntensity = errors.New("invalid intensity")
)

func GetUser(c echo.Context) error {
	id := c.Param("id")

	token, err := redisClient.Get(fmt.Sprintf("%s:access_token", id)).Result()
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	timeNowString := time.Now().Format("2006-01-02")

	activitiesResponse, err := fitbitClient.Activities(id, timeNowString, token)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	userResponse, err := fitbitClient.User(id, token)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	foodGoalsResponse, err := fitbitClient.FoodGoals(token)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	caloriesInResponse, err := fitbitClient.CaloriesIn(timeNowString, token)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	caloriesIn, err := caloriesInResponse.FoodsLogCaloriesIn[0].Calories()
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	calorieDeficitGoal, err := calorieDeficitGoal(foodGoalsResponse.FoodPlan.Intensity)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	caloriesLeftToBurn := caloriesLeftToBurn(calorieDeficitGoal, caloriesIn, activitiesResponse.Summary.CaloriesOut)

	stepsLeftToGo, err := stepsLeftToGo(
		caloriesLeftToBurn,
		caloriesOutPerStep(activitiesResponse.Summary.CaloriesOut, activitiesResponse.Summary.Steps, activitiesResponse.Summary.CaloriesBMR),
	)

	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	getUserResponse := GetUserResponse{
		Avatar:             userResponse.User.Avatar150,
		Name:               userResponse.User.FullName,
		CaloriesIn:         caloriesIn,
		CaloriesOut:        activitiesResponse.Summary.CaloriesOut,
		CalorieDeficitGoal: calorieDeficitGoal,
		StepsGoal:          activitiesResponse.Goals.Steps,
		StepsSoFar:         activitiesResponse.Summary.Steps,
		StepsLeftToGo:      stepsLeftToGo,
		CaloriesLeftToGo:   caloriesLeftToBurn,
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

func caloriesOutPerStep(caloriesOut, stepsSoFar, caloriesBRM int64) float64 {
	caloriesOutFromSteps := caloriesOut - caloriesBRM

	return float64(caloriesOutFromSteps) / float64(stepsSoFar)
}

func caloriesLeftToBurn(calorieDeficitGoal, caloriesIn, caloriesOut int64) int64 {
	return caloriesIn + calorieDeficitGoal - caloriesOut
}

func stepsLeftToGo(caloriesLeftToBurn int64, caloriesOutPerStep float64) (int64, error) {
	return int64(float64(caloriesLeftToBurn) / caloriesOutPerStep), nil
}

func calorieDeficitGoal(intensity string) (int64, error) {
	switch intensity {
	case IntensityMaintenance:
		return IntensityMaintenanceCalories, nil
	case IntensityEasier:
		return IntensityEasierCalories, nil
	case IntensityMedium:
		return IntensityMediumCalories, nil
	case IntensityKindaHard:
		return IntensityKindaHardCalories, nil
	case IntensityHarder:
		return IntensityHarderCalories, nil
	default:
		return 0, ErrInvalidIntensity
	}
}
