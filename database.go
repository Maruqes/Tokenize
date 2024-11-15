package Tokenize

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID       int
	StripeID string
	Email    string
	Name     string
	Password string
}

type DB struct {
	conn *sql.DB
}

func (db *DB) Init() {
	var err error
	db.conn, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}
	db.createTable()
}

func (db *DB) createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		stripe_id TEXT,
		email TEXT,
		name TEXT,
		password TEXT
	);
	`
	_, err := db.conn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func (db *DB) AddUser(stripeID, email, name, password string) (int64, error) {
	hashedPassword := hashPassword(password)
	result, err := db.conn.Exec(`
		INSERT INTO users (stripe_id, email, name, password)
		VALUES (?, ?, ?, ?)
	`, stripeID, email, name, hashedPassword)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *DB) SetUserStripeID(id int, stripeID string) error {
	_, err := db.conn.Exec(`
		UPDATE users
		SET stripe_id = ?
		WHERE id = ?
	`, stripeID, id)
	return err

}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
