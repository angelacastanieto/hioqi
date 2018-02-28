package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/markbates/goth/gothic"
)

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

	return c.JSON(http.StatusOK, activitiesResponse)
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
