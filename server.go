package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"blah": []string{"Welcome to HIOQI"}})
	})
	e.Logger.Fatal(e.Start(":8080"))
}
