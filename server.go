package main

import (
	"os"

	"github.com/angelacastanieto/hioqi/fitbitclient"
	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/fitbit"
)

var (
	err          error
	fitbitClient *fitbitclient.API
	redisClient  *redis.Client
)

func main() {
	fitbitClient, err = fitbitclient.NewAPI()
	if err != nil {
		panic(err)
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	goth.UseProviders(
		fitbit.New(
			os.Getenv("FITBIT_KEY"),
			os.Getenv("FITBIT_SECRET"),
			"http://localhost:8000/auth/fitbit/callback",
			fitbit.ScopeActivity,
			fitbit.ScopeWeight,
			fitbit.ScopeProfile,
			fitbit.ScopeNutrition,
		),
	)

	e := echo.New()
	e.Use(fitbitAuth())
	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,
	}))

	e.GET("/auth/:provider/callback", CallbackHandler)
	e.GET("/auth/:provider", AuthHandler)
	e.GET("/users/:id", GetUser)

	e.Logger.Fatal(e.Start(":8000"))
}

// middleware to build fitbit oauth urls
func fitbitAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			values := c.Request().URL.Query()
			values.Add("provider", "fitbit")
			c.Request().URL.RawQuery = values.Encode()
			return next(c)
		}
	}
}
