package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

type User struct {
	ID       int
	StripeID string
	Email    string
	Name     string
  isProhibited bool
	IsActive bool
}

func CreateTable() {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        stripe_id TEXT,
        email TEXT NOT NULL UNIQUE,
        name TEXT NOT NULL, 
        password TEXT,
    is_prohibited BOOLEAN DEFAULT 0,
		is_active BOOLEAN DEFAULT 0
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

func ProhibitUser(id int) error {
  _, err := db.Exec(`
    UPDATE users
    SET is_prohibited = 1
    WHERE id = ?
  `, id)
  return err
}

func UnprohibitUser(id int) error {
  _, err := db.Exec(`
    UPDATE users
    SET is_prohibited = 0
    WHERE id = ?
  `, id)
  return err
}

func CheckIfUserIsProhibited(id int) (bool, error) {
  row := db.QueryRow(`
    SELECT is_prohibited
    FROM users
    WHERE id = ?
  `, id)
  var isProhibited bool
  err := row.Scan(&isProhibited)
  return isProhibited, err
}


func CheckIfCanUserBeAdded(email, name string) (bool, error) {
	row := db.QueryRow(`
        SELECT id
        FROM users
        WHERE email = ? OR name = ?
    `, email, name)
	var result int
	err := row.Scan(&result)
	if err == sql.ErrNoRows {
		return true, nil // User is unique
	} else if err != nil {
		return false, err // An error occurred
	}
	return false, nil // User exists
}

func AddUser(stripeID, email, name, password string) (int64, error) {
	canBeAdded, err := CheckIfCanUserBeAdded(email, name)
	if err != nil {
		return 0, fmt.Errorf("error checking if user can be added maybe user/email being used")
	}
	if !canBeAdded {
		return 0, fmt.Errorf("user email or username already exists")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return 0, err
	}

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

func CheckIfUserIDExists(id int) (bool, error) {
	row := db.QueryRow(`
		SELECT id
		FROM users
		WHERE id = ?
	`, id)
	var result int
	err := row.Scan(&result)
	if err == sql.ErrNoRows {
		return false, nil // User ID does not exist
	} else if err != nil {
		return false, err // An error occurred
	}
	return true, nil // User ID exists
}

func GetUser(id int) (User, error) {
	row := db.QueryRow(`
		SELECT id, stripe_id, email, name, is_prohibited, is_active
		FROM users
		WHERE id = ?
	`, id)
	var user User
	err := row.Scan(&user.ID, &user.StripeID, &user.Email, &user.Name, &user.isProhibited, &user.IsActive)
	return user, err
}

func GetUserByEmail(email string) (User, error) {
	row := db.QueryRow(`
		SELECT id, stripe_id, email, name, is_prohibited, is_active
		FROM users
		WHERE email = ?
	`, email)
	var user User
	err := row.Scan(&user.ID, &user.StripeID, &user.Email, &user.Name, &user.isProhibited, &user.IsActive)
	return user, err
}

func GetAllUsers() ([]User, error) {
	rows, err := db.Query(`
		SELECT id, stripe_id, email, name,is_prohibited, is_active
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.StripeID, &user.Email, &user.Name, &user.isProhibited, &user.IsActive)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func ActivateUser(id int) error {
	_, err := db.Exec(`
		UPDATE users
		SET is_active = 1
		WHERE id = ?
	`, id)
	return err
}

func DeactivateUser(id int) error {
	_, err := db.Exec(`
		UPDATE users
		SET is_active = 0
		WHERE id = ?
	`, id)
	return err
}

func DeactivateUserByStripeID(stripeID string) error {
	_, err := db.Exec(`
		UPDATE users
		SET is_active = 0
		WHERE stripe_id = ?
	`, stripeID)
	return err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CheckUserPassword(id int, password string) bool {
	row := db.QueryRow(`
		SELECT password
		FROM users
		WHERE id = ?
	`, id)
	var hashedPassword string
	err := row.Scan(&hashedPassword)
	if err != nil {
		return false
	}
	return VerifyPassword(password, hashedPassword)
}
