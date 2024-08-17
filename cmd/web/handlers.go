package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Tsundere-Musume/message/internal/models"
	"github.com/Tsundere-Musume/message/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "home.html", data)
}

func (app *application) userSignUp(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = SignUpForm{}
	app.render(w, http.StatusOK, "signup.html", data)
}

func (app *application) userSignUpPost(w http.ResponseWriter, r *http.Request) {
	var form SignUpForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "Name can't be empty.")
	form.CheckField(validator.NotBlank(form.Email), "email", "Email can't be empty.")
	form.CheckField(validator.NotBlank(form.Password), "password", "Password can't be empty.")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Invalid email.")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long.")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	filename, err := app.uploadAvatar(r)
	fmt.Println(filename)
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			app.serverErrror(w, err)
		}
		return
	}
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogIn(w http.ResponseWriter, r *http.Request) {
	form := LogInForm{}
	data := app.newTemplateData(r)
	data.Form = form
	app.render(w, http.StatusOK, "login.html", data)
}

func (app *application) userLogInPost(w http.ResponseWriter, r *http.Request) {
	var form LogInForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "Email can't be empty.")
	form.CheckField(validator.NotBlank(form.Password), "password", "Password can't be empty.")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Invalid email.")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect.")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
			return
		} else {
			app.serverErrror(w, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverErrror(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id.String())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogOutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverErrror(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userList(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	user, err := app.users.Get(userId)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverErrror(w, err)
		}
		return
	}

	users, err := app.users.GetAllUsers(userId)
	if err != nil {
		app.serverErrror(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.User = user
	data.Users = users

	app.render(w, http.StatusOK, "user_list.html", data)
}

func (app *application) friendList(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	user, err := app.users.Get(userId)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverErrror(w, err)
		}
		return
	}

	users, err := app.users.GetFriends(userId)
	if err != nil {
		app.serverErrror(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.User = user
	data.Users = users

	app.render(w, http.StatusOK, "user_list.html", data)
}

func (app *application) addFriend(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	//get it from a post form
	params := httprouter.ParamsFromContext(r.Context())
	otherUserID := params.ByName("id")
	if otherUserID == "" {
		app.notFound(w)
		return
	}

	err := app.users.AddFriend(userId, otherUserID)
	if err != nil {
		//TODO: handle errors for constraint better
		app.clientError(w, http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (app *application) removeFriend(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	params := httprouter.ParamsFromContext(r.Context())
	otherUserID := params.ByName("id")
	if otherUserID == "" {
		app.notFound(w)
		return
	}

	err := app.users.RemoveFriend(userId, otherUserID)
	if err != nil {
		//TODO: handle errors for constraint better
		app.clientError(w, http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
