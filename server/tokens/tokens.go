package tokens

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
)

var (
	headersCSV      = []string{"login", "password", "token"}
	lettersForToken = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

type Tokens struct {
	m     sync.RWMutex
	users [][]string
}

func New() *Tokens {
	return &Tokens{}
}

func NewFrom(reader io.Reader) *Tokens {

	var result Tokens

	r := csv.NewReader(reader)
	records, err := r.ReadAll()
	if err != nil {
		panic(fmt.Errorf("failed to open reader %s", err))
	}

	for _, eachrecord := range records {
		if eachrecord[0] != "login" && eachrecord[1] != "password" && eachrecord[2] != "token" {
			result.users = append(result.users, eachrecord)
		}
	}

	return &result
}

func NewFromFile(path string) *Tokens {

	var result Tokens

	reader, _ := os.Open(path)
	defer reader.Close()

	r := csv.NewReader(reader)
	records, err := r.ReadAll()
	if err != nil {
		panic(fmt.Errorf("failed to open file %s", err))
	}

	for _, eachrecord := range records {
		if eachrecord[0] != "login" && eachrecord[1] != "password" && eachrecord[2] != "token" {
			result.users = append(result.users, eachrecord)
		}
	}

	return &result
}

func (t *Tokens) WriteTo(writer io.Writer) (int64, error) {
	var records [][]string
	records = append(records, headersCSV)
	records = append(records, t.users...)

	w := csv.NewWriter(writer)
	w.WriteAll(records)
	if err := w.Error(); err != nil {
		return 0, err
	}

	return int64(len(records)), nil
}

func (t *Tokens) AddUser(login string, passwod string) error {

	t.m.RLock()
	for _, item := range t.users {
		if item[0] == login {
			return fmt.Errorf("the user %s already exists", login)
		}
	}
	t.m.RUnlock()
	var newUser []string
	newUser = append(newUser, login, passwod, randString(15))

	t.m.Lock()
	defer t.m.Unlock()
	t.users = append(t.users, newUser)

	return nil
}

func (t *Tokens) GetPassword(login string) (string, bool) {

	t.m.RLock()
	defer t.m.RUnlock()

	for _, item := range t.users {
		if item[0] == login {
			return item[1], true
		}
	}
	return "", false
}

func (t *Tokens) DeleteUser(login string) {
	var indexUser int = -1
	t.m.RLock()
	for index, item := range t.users {
		if item[0] == login {
			indexUser = index
			break
		}
	}
	t.m.RUnlock()

	if indexUser > -1 {
		t.m.Lock()
		t.users = append(t.users[:indexUser], t.users[indexUser+1:]...)
		t.m.Unlock()
	}
}

func (t *Tokens) GetToken(login string) (string, error) {
	t.m.RLock()
	defer t.m.RUnlock()
	for _, item := range t.users {
		if item[0] == login {
			return item[2], nil
		}
	}
	return "", fmt.Errorf("user with login %s not exist", login)
}

func (t *Tokens) ValidToken(token string) (bool, error) {
	t.m.RLock()
	defer t.m.RUnlock()
	for _, item := range t.users {
		if item[2] == token {
			return true, nil
		}
	}
	return false, fmt.Errorf("token does not exist")
}

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = lettersForToken[rand.Intn(len(lettersForToken))]
	}
	return string(b)
}
