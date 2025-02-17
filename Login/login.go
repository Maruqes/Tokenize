package Login

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Maruqes/Tokenize/database"
)

type Login struct {
	UserID  int
	Token   string
	Expires int64
}

type LoginStore struct {
	sync.RWMutex
	logins map[int]Login
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func checkForExpiredLogins() {
	const expirationDays = 7
	const checkInterval = time.Hour

	for {
		time.Sleep(checkInterval)
		loginStore.Lock()
		for id, login := range loginStore.logins {
			loginTime := time.Unix(login.Expires, 0)
			if loginTime.Add(time.Hour * 24 * expirationDays).Before(time.Now()) {
				delete(loginStore.logins, id)
			}
		}
		loginStore.Unlock()
	}
}

func newLoginStore() *LoginStore {

	return &LoginStore{
		logins: make(map[int]Login),
	}
}

var loginStore = newLoginStore()

func (s *LoginStore) add(login Login) {
	s.Lock()
	defer s.Unlock()
	login.Expires = time.Now().Unix()
	s.logins[login.UserID] = login
}

func (s *LoginStore) get(userID int) (Login, bool) {
	s.RLock()
	defer s.RUnlock()
	login, ok := s.logins[userID]
	return login, ok
}

func (s *LoginStore) delete(userID int) {
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

	loginStore.add(Login{
		UserID: usr.ID,
		Token:  token,
	})
	return token, usr, nil
}

func LogoutUser(userID int) {
	loginStore.delete(userID)
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
	login, ok := loginStore.get(id)
	if !ok || login.Token != token {
		return false
	}
	return true
}

func GetIdWithRequest(r *http.Request) (int, error) {
	//get cookies id and token
	cookie, err := r.Cookie("id")
	if err != nil {
		return -1, err
	}
	id, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func Init() {
	go checkForExpiredLogins()
}
