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

	fileServer := http.FileServer(http.Dir("./ui/static"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	dynamic.ThenFunc(app.home)
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignUp))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignUpPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogIn))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLogInPost))

	protected := dynamic.Append(app.requireAuthentication)
	router.Handler(http.MethodGet, "/message/:id", protected.ThenFunc(app.directMessage))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogOutPost))
	router.Handler(http.MethodGet, "/chat", protected.ThenFunc(app.userList))
	router.Handler(http.MethodGet, "/subscribe/:id", protected.ThenFunc(app.subscriberHandler))
	router.Handler(http.MethodPost, "/publish", protected.ThenFunc(app.directMessagePost))
	router.Handler(http.MethodGet, "/user/add/:id", protected.ThenFunc(app.addFriend))
	router.Handler(http.MethodGet, "/user/remove/:id", protected.ThenFunc(app.removeFriend))
	base := alice.New(app.logRequest, secureHeaders)
	return base.Then(router)
}
