package main

import (
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/markbates/goth/gothic"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	view.HTML(w, 200, "index", nil)
}

func getUser(w http.ResponseWriter, r *http.Request) {

	view.HTML(w, 200, "index", nil)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	session, _ := store.Get(r, user.UserID)

	session.Values["access_token"] = user.AccessToken
	session.Values["refresh_token"] = user.RefreshToken

	// Save it before we write to the response/return from the handler.
	session.Save(r, w)

	spew.Dump(user) // Store USER KEYS HERE
	http.Redirect(w, r, "http://localhost:3000/users/"+user.UserID, http.StatusSeeOther)

	// view.HTML(w, 200, "user", user)
}
