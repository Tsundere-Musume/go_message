package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Tsundere-Musume/message/internal/models"
	"github.com/julienschmidt/httprouter"
	"nhooyr.io/websocket"
)

func (app *application) directMessage(w http.ResponseWriter, r *http.Request) {
	senderId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	params := httprouter.ParamsFromContext(r.Context())
	receiverId := params.ByName("id")
	if receiverId == "" {
		app.notFound(w)
		return
	}

	recv, err := app.users.Get(receiverId)
	if err != nil {
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
	data.Heading = recv.Name
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
	user, err := app.users.Get(senderId)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverErrror(w, err)
		}
		return
	}
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
		Sender:  user.Name,
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
