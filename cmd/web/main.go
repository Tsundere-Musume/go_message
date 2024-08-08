package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Tsundere-Musume/message/internal/models"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/lib/pq"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	templates      map[string]*template.Template
	formDecoder    *form.Decoder
	users          *models.UserModel
	directMessages *models.DirectMessageModel
	sessionManager *scs.SessionManager
	// chat                *chatRoom
	directMessageServer *directMsgServer
}

func main() {
	addr := ":4000"
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	templates, err := newTemplateCache()
	if err != nil {
		errLog.Fatal(err)
	}

	db, err := openDB("postgres://postgres:db123@localhost:5432/message?sslmode=disable")
	if err != nil {
		errLog.Fatalln(err)
	}

	defer db.Close()

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = false

	app := application{
		templates:      templates,
		errorLog:       errLog,
		infoLog:        infoLog,
		users:          &models.UserModel{DB: db},
		formDecoder:    form.NewDecoder(),
		sessionManager: sessionManager,
		// chat:                newChatServer(),
		directMessages:      &models.DirectMessageModel{DB: db},
		directMessageServer: serverDM(),
	}

	// TODO: add https
	srv := &http.Server{
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 10,
		Handler:      app.routes(),
		Addr:         addr,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	infoLog.Printf("Started server on %v", addr)

	// TODO: add graceful shutdown
	err = <-errCh
	errLog.Fatalf("Error while running the server: %s", err)
}
func openDB(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(); err != nil {
		return nil, err
	}
	err = models.InitUsers(conn)
	if err != nil {
		return nil, err
	}

	err = models.InitSession(conn)
	if err != nil {
		return nil, err
	}

	err = models.InitDirectMessage(conn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
