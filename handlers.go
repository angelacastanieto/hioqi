package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/markbates/goth/gothic"
)

func GetUser(c echo.Context) error {
	id := c.Param("id")
	fmt.Println("ID", id)

	sess, err := session.Get("sessions", c)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	// Retrieve our access_token and type-assert it
	token, ok := sess.Values["access_token"].(string)

	if !ok {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"errors": []string{"token not found"}})
	}

	activities, err := fitbitClient.Activities(id, time.Now().Format("2006-01-02"), token)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"errors": []string{err.Error()}})
	}

	spew.Dump(activities)
	return c.JSON(http.StatusOK, map[string]interface{}{"activities": activities})
}

func AuthHandler(c echo.Context) error {
	gothic.BeginAuthHandler(c.Response(), c.Request())

	return nil
}

func CallbackHandler(c echo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		fmt.Println(err)
		http.Redirect(c.Response().Writer, c.Request(), "http://localhost:3000", http.StatusInternalServerError)
	}

	sess, err := session.Get("sessions", c)
	if err != nil {
		fmt.Println(err)
		http.Redirect(c.Response().Writer, c.Request(), "http://localhost:3000", http.StatusInternalServerError)
	}

	sess.Values["access_token"] = user.AccessToken
	sess.Values["refresh_token"] = user.RefreshToken

	err = sess.Save(c.Request(), c.Response())
	fmt.Println("ERRR IS", err)
	if err != nil {
		fmt.Println(err)
		http.Redirect(c.Response().Writer, c.Request(), "http://localhost:3000", http.StatusInternalServerError)
	}

	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000/users/"+user.UserID)
}
