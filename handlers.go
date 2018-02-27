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

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	spew.Dump(user)

	view.HTML(w, 200, "user", user)
}
