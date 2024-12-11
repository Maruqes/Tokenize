package UserFuncs

import "github.com/Maruqes/Tokenize/database"

func GetAllUsers() ([]database.User, error) {
	return database.GetAllUsers()
}

func GetUserByID(id int) (database.User, error) {
	return database.GetUser(id)
}

func GetUserByEmail(email string) (database.User, error) {
	return database.GetUserByEmail(email)
}
