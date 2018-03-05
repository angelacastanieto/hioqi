package main

import (
	"fmt"
	"os"

	"github.com/angelacastanieto/hioqi/fitbitclient"
	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/fitbit"
	redistore "gopkg.in/boj/redistore.v1"
)

var (
	err          error
	fitbitClient *fitbitclient.API
	redisClient  *redis.Client
	store        *redistore.RediStore
	env          = "development"
	appConfig    Config
)

type Config struct {
	Host string
	Port string
}

func main() {
	if os.Getenv("ENVIRONMENT") != "" { // need to use config package
		env = os.Getenv("ENVIRONMENT")
	}

	appConfig = config(env)

	store, err = redistore.NewRediStore(16, "tcp", os.Getenv("REDIS_URL"), os.Getenv("REDIS_PASSWORD"), []byte("secret-key"))
	if err != nil {
		panic(err)
	}

	store.Options = &sessions.Options{
		Path:     "/",      // to match all requests
		MaxAge:   3600 * 1, // 1 hour
		HttpOnly: true,
	}

	fitbitClient, err = fitbitclient.NewAPI()
	if err != nil {
		panic(err)
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	goth.UseProviders(
		fitbit.New(
			os.Getenv("FITBIT_KEY"),
			os.Getenv("FITBIT_SECRET"),
			fmt.Sprintf("%s/auth/fitbit/callback", appConfig.Host),
			fitbit.ScopeActivity,
			fitbit.ScopeWeight,
			fitbit.ScopeProfile,
			fitbit.ScopeNutrition,
		),
	)

	e := echo.New()
	e.Use(fitbitAuth())
	e.Use(session.Middleware(store))

	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,
	}))

	e.GET("/auth/:provider/callback", CallbackHandler)
	e.GET("/auth/:provider", AuthHandler)
	e.GET("/users/:id", GetUser)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", appConfig.Port)))
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

// use for now until get config package
func config(env string) Config {
	var config Config
	if env == "production" {
		config.Host = "https://floating-depths-67623.herokuapp.com/"
		config.Port = os.Getenv("PORT")
	} else {
		config.Host = "http://localhost:8000"
		config.Port = "8000"
	}
	return config
}
