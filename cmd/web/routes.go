package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	dynamic.ThenFunc(app.home)
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignUp))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignUpPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogIn))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLogInPost))

	protected := dynamic.Append(app.requireAuthentication)
	// TODO: change it to post
	router.Handler(http.MethodGet, "/user/logout", protected.ThenFunc(app.userLogOutPost))

	base := alice.New(app.logRequest, secureHeaders)
	return base.Then(router)
}
