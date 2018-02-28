package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/markbates/goth/gothic"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["id"]

	if !ok || len(keys) < 1 {
		log.Println("Url Param 'id' is missing")
		return
	}

	id := keys[0]

	session, _ := store.Get(r, id)

	// Retrieve our access_token and type-assert it
	token, ok := session.Values["access_token"].(string)

	if !ok {
		return
	}

	activities, err := fitbitClient.Activities(id, time.Now().Format("2006-01-02"), token)
	if err != nil {
		return
	}

	spew.Dump(activities)
	json.NewEncoder(w).Encode(r)
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
}
