package main

import (
	"os"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/fitbit"
	render "gopkg.in/unrolled/render.v1"
)

var view = render.New(render.Options{
	Directory:     "templates",
	Extensions:    []string{".html"},
	IsDevelopment: true,
	IndentJSON:    true,
})

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

func init() {
	goth.UseProviders(
		fitbit.New(
			os.Getenv("FITBIT_KEY"),
			os.Getenv("FITBIT_SECRET"),
			"http://localhost:3000/auth/fitbit/callback",
			fitbit.ScopeActivity,
			fitbit.ScopeWeight,
			fitbit.ScopeProfile,
			fitbit.ScopeNutrition,
		),
	)
}

func main() {
	p := pat.New()

	p.Get("/auth/{provider}/callback", CallbackHandler)
	p.Get("/auth/{provider}", gothic.BeginAuthHandler)
	p.Get("/", IndexHandler)

	n := negroni.Classic()
	n.UseHandler(p)
	n.Run(":3000")
}
