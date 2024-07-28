package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword string
	CreatedAt      time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `
	INSERT INTO users (id, name, email, hashed_password, created) 
	VALUES ($1, $2, $3, $4, $5);
	`
	_, err = m.DB.Exec(stmt, uuid.New(), name, email, hashedPassword, time.Now().UTC())
	if err != nil {
		if strings.Contains(err.Error(), "users_email_key") {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (uuid.UUID, error) {
	var id uuid.UUID
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = $1"
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		//TODO: use better driver that returns has errors types maybe
		if strings.Contains(err.Error(), "no rows in result set") {
			return uuid.UUID{}, ErrInvalidCredentials
		}
		return uuid.UUID{}, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return uuid.UUID{}, ErrInvalidCredentials
		}
		return uuid.UUID{}, err
	}

	return id, nil
}

func (m *UserModel) Exists(id string) (bool, error) {
	var exists bool

	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id = $1)"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err

}

func (m *UserModel) Get(id string) (*User, error) {
	var user User
	stmt := "SELECT name, email, created FROM users WHERE id = $1"
	err := m.DB.QueryRow(stmt, id).Scan(&user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return &user, nil
}
