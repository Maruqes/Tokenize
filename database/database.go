package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func CreateTable() {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        stripe_id TEXT,
        email TEXT,
        name TEXT,
        password TEXT
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func Init() {
	var err error
	// Check if the database file exists
	if _, err := os.Stat("./users.db"); os.IsNotExist(err) {
		// Create the database file if it does not exist
		file, err := os.Create("./users.db")
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}

	db, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}
	CreateTable()
}

func AddUser(stripeID, email, name, password string) (int64, error) {
	hashedPassword := hashPassword(password)
	result, err := db.Exec(`
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

func SetUserStripeID(id int, stripeID string) error {
	_, err := db.Exec(`
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
