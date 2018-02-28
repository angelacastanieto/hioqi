package main

import (
	"os"

	"github.com/angelacastanieto/hioqi/fitbitclient"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/fitbit"
)

var (
	err          error
	fitbitClient *fitbitclient.API
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

func init() {
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
}

func main() {
	fitbitClient, err = fitbitclient.NewAPI()
	if err != nil {
		panic(err)
	}

	p := pat.New()

	p.Get("/auth/{provider}/callback", CallbackHandler)
	p.Get("/auth/{provider}", gothic.BeginAuthHandler)
	p.Get("/users/{id}", GetUser)

	n := negroni.Classic()
	n.UseHandler(p)
	n.Run(":8000")
}
