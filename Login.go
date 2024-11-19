package Tokenize

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"

	"github.com/Maruqes/Tokenize/database"
)

type Login struct {
	UserID int
	Token  string
}

var logins = map[int]Login{}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func generateSecureToken(length int) (string, error) {
	token := make([]byte, length)
	for i := range token {
		// Escolhe um índice aleatório dentro do charset
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		token[i] = charset[num.Int64()]
	}
	return string(token), nil
}

func loginUser(email, password string) (string, database.User, error) {
	usr, err := database.GetUserByEmail(email)
	if err != nil {
		return "", usr, err
	}

	login := database.CheckUserPassword(usr.ID, password)
	if !login {
		return "", usr, nil
	}

	token, err := generateSecureToken(64)
	if err != nil {
		return "", usr, err
	}

	logins[usr.ID] = Login{
		UserID: usr.ID,
		Token:  token,
	}
	return token, usr, nil
}

func logoutUser(userID int) {
	delete(logins, userID)
}

func checkToken(r *http.Request) bool {
	//get cookies id and token
	cookie, err := r.Cookie("id")
	if err != nil {
		return false
	}
	id, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return false
	}
	cookie, err = r.Cookie("token")
	if err != nil {
		return false
	}
	token := cookie.Value

	//check if token is valid
	login, ok := logins[id]
	if !ok {
		return false
	}
	if login.Token != token {
		return false
	}
	return true
}
