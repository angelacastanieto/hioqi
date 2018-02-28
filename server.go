package main

import (
	"os"

	"github.com/angelacastanieto/hioqi/fitbitclient"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/fitbit"
)

var (
	err          error
	fitbitClient *fitbitclient.API
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

func init() {
	store.Options = &sessions.Options{
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
	}
}

func main() {
	fitbitClient, err = fitbitclient.NewAPI()
	if err != nil {
		panic(err)
	}

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
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
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
