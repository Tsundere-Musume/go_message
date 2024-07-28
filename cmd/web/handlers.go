package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/Tsundere-Musume/message/internal/models"
	"github.com/Tsundere-Musume/message/internal/validator"
	"nhooyr.io/websocket"
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

func (app *application) getChatPage(w http.ResponseWriter, r *http.Request) {
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
	data := app.newTemplateData(r)
	data.User = user
	app.render(w, http.StatusOK, "message.html", data)
}

func (app *application) subscriberHandler(w http.ResponseWriter, r *http.Request) {
	err := app.chat.subscribe(r.Context(), w, r)
	if errors.Is(err, context.Canceled) {
		return
	}

	if websocket.CloseStatus(err) == websocket.StatusNormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}

	if err != nil {
		//TODO: change the webpage somehow
		app.errorLog.Println(err)
		return
	}
}

func (app *application) publishMessage(w http.ResponseWriter, r *http.Request) {
	// body := http.MaxBytesReader(w, r.Body, 8192)
	// fmt.Println(r.Body)
	// msg_content, err := io.ReadAll(body)
	// if err != nil {
	// 	app.clientError(w, http.StatusRequestEntityTooLarge)
	// 	return
	// }

	//TODO:
	// maybe add the data into the body instead of a form
	// check the cost of processing body vs form requests

	var form MessageForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

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

	msg := message{
		UserID:   userId,
		Username: user.Name,
		Value:    form.Message,
	}
	app.chat.publish(msg)
	w.WriteHeader(http.StatusAccepted)
}
