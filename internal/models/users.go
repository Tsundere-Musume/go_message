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
	ID             uuid.UUID
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

func (m *UserModel) GetAllUsers(id string) ([]*User, error) {
	//TODO: maybe remove the email probably should remove better to remove
	stmt := "SELECT id, name, email, created FROM users WHERE id != $1"
	rows, err := m.DB.Query(stmt, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	users := []*User{}

	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (m *UserModel) AddFriend(userId, otherId string) error {
	stmt := `
    INSERT INTO FRIENDS (user_id_1, user_id_2)
    VALUES ($1, $2)
    ON CONFLICT (user_id_1, user_id_2) DO NOTHING;
  `
	var err error
	if userId < otherId {
		_, err = m.DB.Exec(stmt, userId, otherId)
	} else {
		_, err = m.DB.Exec(stmt, otherId, userId)
	}
	return err
}

func (m *UserModel) RemoveFriend(userId, otherId string) error {
	stmt := `
    DELETE FROM FRIENDS 
    WHERE user_id_1 = $1 AND user_id_2 = $2;
  `
	var err error
	if userId < otherId {
		_, err = m.DB.Exec(stmt, userId, otherId)
	} else {
		_, err = m.DB.Exec(stmt, otherId, userId)
	}
	return err
}
