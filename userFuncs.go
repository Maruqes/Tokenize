package Tokenize

import "github.com/Maruqes/Tokenize/database"

type Users struct {
}

func (*Users) GetAllUsers() ([]database.User, error) {
	return database.GetAllUsers()
}

func (*Users) GetUserByID(id int) (database.User, error) {
	return database.GetUser(id)
}

func (*Users) GetUserByEmail(email string) (database.User, error) {
	return database.GetUserByEmail(email)
}
