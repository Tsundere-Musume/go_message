package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Tsundere-Musume/message/internal/models"
	"github.com/go-playground/form/v4"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (app *application) serverErrror(w http.ResponseWriter, err error) {
	app.errorLog.Println(err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	tmpl, ok := app.templates[page]
	if !ok {
		err := fmt.Errorf("template %s does not exist", page)
		app.serverErrror(w, err)
		return
	}

	w.WriteHeader(status)
	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverErrror(w, err)
	}
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	return &templateData{
		UserID:          userID,
		CSRFToken:       nosurf.Token(r),
		IsAuthenticated: app.isAuthenticated(r),
	}
}

func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
		return err
	}
	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

// func serializeMessage(msg message) ([]byte, error) {
// 	val, err := json.Marshal(msg)
// 	if err != nil {
// 		// do something with this error
// 		return []byte{}, nil
// 	}
// 	return val, nil
// }
//
// // NOTE:  Currently not used
// func deserializeMessage(b []byte) (message, error) {
// 	var msg message
// 	err := json.Unmarshal(b, &msg)
// 	if err != nil {
// 		return msg, err
// 	}
// 	return msg, nil
// }

func (app *application) uploadAvatar(r *http.Request) (string, error) {
	err := r.ParseMultipartForm(10 << 20) //TODO: CHANGE THE UPLOAD SIZE
	if err != nil {
		return "", err
	}

	file, fileHeaders, err := r.FormFile("avatar")
	if err != nil {
		return "", models.ErrNoAvatarImg
	}

	defer file.Close()

	fileExt := filepath.Ext(fileHeaders.Filename)
	filename := uuid.NewString() + fileExt
	outFile, err := os.Create("./ui/static/images/" + filename)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return "", err
	}
	return filename, nil
}
