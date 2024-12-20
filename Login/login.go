package Login

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	functions "github.com/Maruqes/Tokenize/Functions"
	"github.com/Maruqes/Tokenize/database"
	"github.com/Maruqes/Tokenize/offline"
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

func newLoginStore() *LoginStore {
	return &LoginStore{
		logins: make(map[int]Login),
	}
}

var loginStore = newLoginStore()

func (s *LoginStore) add(login Login) {
	s.Lock()
	defer s.Unlock()
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

func IsUserActiveRequest(r *http.Request) (bool, error) {
	cookie, err := r.Cookie("id")
	if err != nil {
		return false, fmt.Errorf("error getting id")
	}
	id, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return false, fmt.Errorf("error converting id to int")
	}
	cookie, err = r.Cookie("token")
	if err != nil {
		return false, fmt.Errorf("error getting token")
	}
	token := cookie.Value

	//check if token is valid
	login, ok := loginStore.get(id)
	if !ok || login.Token != token {
		return false, fmt.Errorf("invalid token")
	}

	if off, err := offline.IsAccountActivatedOffline(id); 
		time.Now().Unix() < time.Date(off.End_date.Year, time.Month(off.End_date.Month), off.End_date.Day, 0, 0, 0, 0, time.UTC).Unix() || err != nil {
		return true, nil
	}

	return functions.DoesUserHaveActiveSubscription(id)
}

func IsUserActive(id int) (bool, error) {
	return functions.DoesUserHaveActiveSubscription(id)
}
