package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/markbates/goth/gothic"
)

// TODO: figure out case where you havent taken any steps yet that day
// TODO; what if you have reached your goal or gone over your goal for deficit?

type (
	GetUserResponse struct {
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

	Something struct {
		UserID string `json:"user_id"`
	}
)

func GetUser(c echo.Context) error {
	id := c.Param("id")
	var resync bool
	resyncString := c.QueryParam("resync")
	sess, err := session.Get("user_session", c)
	if err != nil {
		fmt.Println(err)
		return c.Redirect(http.StatusTemporaryRedirect, appConfig.HioqiWebURL)
	}

	loggedInUser, ok := sess.Values["user_id"]
	if !ok {
		return c.Redirect(http.StatusTemporaryRedirect, appConfig.HioqiWebURL)
	}

	if loggedInUser != id {
		fmt.Println(loggedInUser, "unauthorized for", id)
		return c.NoContent(http.StatusUnauthorized)
	}

	if resyncString != "" {
		resync, err = strconv.ParseBool(resyncString)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"errors": []string{err.Error()}})
		}
	}

	token, ok := sess.Values["access_token"]
	if !ok {
		fmt.Println("No access token", err)
		return c.Redirect(http.StatusTemporaryRedirect, appConfig.HioqiWebURL)
	}

	if !resync {
		userResponseCachedJSON, err := redisClient.Get(fmt.Sprintf("%s:user_response", id)).Result()
		if err != nil {
			fmt.Println(err)
		}

		if userResponseCachedJSON == "" {
			fmt.Println("nothing cached - will make new request")
		} else {
			var getUserResponse GetUserResponse
			err = json.Unmarshal([]byte(userResponseCachedJSON), &getUserResponse)
			if err != nil { // if err, log err and get new userResponse from fitbit
				fmt.Println("Redis JSON unmarshaling error", err)
			} else {
				fmt.Println("returning cached data")
				return c.JSON(http.StatusOK, getUserResponse)
			}
		}
	}

	fmt.Println("getting new data")

	timeNowString := time.Now().Format("2006-01-02")

	activitiesResponse, err := fitbitClient.Activities(id, timeNowString, token.(string))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	userResponse, err := fitbitClient.User(id, token.(string))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	foodGoalsResponse, err := fitbitClient.FoodGoals(token.(string))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	caloriesInResponse, err := fitbitClient.CaloriesIn(timeNowString, token.(string))
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	caloriesIn, err := caloriesInResponse.FoodsLogCaloriesIn[0].Calories()
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	calorieDeficitGoal, err := foodGoalsResponse.FoodPlan.CalorieDeficitGoal()
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

	getUserResponseBytes, err := json.Marshal(getUserResponse)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	err = redisClient.Set(fmt.Sprintf("%s:user_response", id), string(getUserResponseBytes[:]), time.Minute*20).Err()
	if err != nil {
		fmt.Println(err)
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
		return c.Redirect(http.StatusTemporaryRedirect, appConfig.HioqiWebURL)
	}

	sess, err := session.Get("user_session", c)
	if err != nil {
		fmt.Println("get session error", err)
		return c.Redirect(http.StatusTemporaryRedirect, appConfig.HioqiWebURL)
	}

	sess.Values["user_id"] = user.UserID
	sess.Values["access_token"] = user.AccessToken
	sess.Values["refresh_token"] = user.RefreshToken

	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/users/%s", appConfig.HioqiWebURL, user.UserID))
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
