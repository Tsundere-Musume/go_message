package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Tsundere-Musume/message/internal/models"
	"github.com/Tsundere-Musume/message/internal/validator"
	"github.com/julienschmidt/httprouter"
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

func (app *application) directMessage(w http.ResponseWriter, r *http.Request) {
	senderId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	params := httprouter.ParamsFromContext(r.Context())
	receiverId := params.ByName("id")
	if receiverId == "" {
		app.notFound(w)
		return
	}
	messages, err := app.directMessages.GetMessagesForUser(senderId, receiverId)
	if err != nil {
		app.serverErrror(w, err)
		return
	}
	data := app.newTemplateData(r)
	data.Messages = messages
	app.render(w, http.StatusOK, "message.html", data)

}

func (app *application) subscriberHandler(w http.ResponseWriter, r *http.Request) {
	senderId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	//TODO: validate id
	ctx := context.WithValue(r.Context(), userIdKey, senderId)
	r = r.WithContext(ctx)
	err := app.directMessageServer.subscribe(r.Context(), w, r)
	if errors.Is(err, context.Canceled) {
		return
	}

	if websocket.CloseStatus(err) == websocket.StatusNormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}

	if err != nil {
		//TODO: HANDLE ERRORS
		if errors.Is(err, models.ErrNotFound) {
			app.notFound(w)
			return
		}
		//TODO: change the webpage somehow
		//	dont understand what this todo was supposed to meant
		app.errorLog.Println(err)
		return
	}
}

func (app *application) directMessagePost(w http.ResponseWriter, r *http.Request) {
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
	var form DirectMessageForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	senderId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	err = app.directMessages.Send(senderId, form.ReceiverID, form.Message)
	if err != nil {
		app.serverErrror(w, err)
		return
	}
	msg := models.DirectMessage{
		FromId:  senderId,
		ToId:    form.ReceiverID,
		Body:    form.Message,
		Created: time.Now().UTC(),
	}
	//validate user id
	room, ok := app.directMessageServer.getRoomByIds(msg.FromId, msg.ToId)
	if !ok {
		app.notFound(w)
		return
	}
	room.publish(msg)
	w.WriteHeader(http.StatusAccepted)
}
