package models

import "database/sql"

func InitUsers(db *sql.DB) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS USERS(
	id UUID	PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	hashed_password CHAR(60) NOT NULL,
	created TIMESTAMP NOT NULL
	);


  CREATE TABLE IF NOT EXISTS friends(
  user_id_1 UUID,
  user_id_2 UUID,
  PRIMARY KEY (user_id_1, user_id_2),
  FOREIGN KEY (user_id_1) REFERENCES USERS(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id_2) REFERENCES USERS(id) ON DELETE CASCADE,
  CHECK (user_id_1 <> user_id_2)
  );
	`
	_, err := db.Exec(stmt)
	return err
}

func InitSession(db *sql.DB) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
	);

	CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions (expiry);
	`
	_, err := db.Exec(stmt)
	return err
}

func InitDirectMessage(db *sql.DB) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS direct_message (
	from_id UUID REFERENCES users(id) ON DELETE CASCADE,
	to_id UUID REFERENCES users(id) ON DELETE CASCADE,
	body TEXT NOT NULL,
	created TIMESTAMP NOT NULL
	);`
	_, err := db.Exec(stmt)
	return err
}
