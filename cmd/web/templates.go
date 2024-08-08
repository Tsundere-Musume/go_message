package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/Tsundere-Musume/message/internal/models"
)

type templateData struct {
	Form            any
	UserID          string
	User            *models.User
	Users           []*models.User
	Messages        []*models.DirectMessage // TODO: change it to a more generic message type later or add two separate messages for direct message or group message
	IsAuthenticated bool
	CSRFToken       string
}

func chatTime(t time.Time, timezone string) string {
	//WARN:
	//TODO: Fix time according to user thing timezone ??
	if t.IsZero() {
		return ""
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t.UTC().Format("03:04:05 PM")
	}
	return t.In(loc).Format("03:04:05 PM")
}

var functions = template.FuncMap{
	"chatTime": chatTime,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)
	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}
