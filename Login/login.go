package Login

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"sync"

	"github.com/Maruqes/Tokenize/database"
)

type Login struct {
	UserID int
	Token  string
}

type LoginStore struct {
	sync.RWMutex
	logins map[int]Login
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func NewLoginStore() *LoginStore {
	return &LoginStore{
		logins: make(map[int]Login),
	}
}

var loginStore = NewLoginStore()

func (s *LoginStore) Add(login Login) {
	s.Lock()
	defer s.Unlock()
	s.logins[login.UserID] = login
}

func (s *LoginStore) Get(userID int) (Login, bool) {
	s.RLock()
	defer s.RUnlock()
	login, ok := s.logins[userID]
	return login, ok
}

func (s *LoginStore) Delete(userID int) {
	s.Lock()
	defer s.Unlock()
	delete(s.logins, userID)
}

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

func LoginUser(email, password string) (string, database.User, error) {
	usr, err := database.GetUserByEmail(email)
	if err != nil {
		return "", usr, err
	}

	login := database.CheckUserPassword(usr.ID, password)
	if !login {
		return "", usr, fmt.Errorf("invalid password or user")
	}

	token, err := generateSecureToken(64)
	if err != nil {
		return "", usr, err
	}

	loginStore.Add(Login{
		UserID: usr.ID,
		Token:  token,
	})
	return token, usr, nil
}

func LogoutUser(userID int) {
	loginStore.Delete(userID)
}

func CheckToken(r *http.Request) bool {
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
	login, ok := loginStore.Get(id)
	if !ok || login.Token != token {
		return false
	}

	return true
}
