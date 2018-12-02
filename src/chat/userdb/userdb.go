package userdb

import (
	"errors"
	"fmt"
	"math/rand"
)

type UserDb interface {
	MakeNewGuestName() string
	IsRegistered(name string) bool
	Register(name string, password string) error
	Authenticate(name string, password string) error
}

type InMemoryUserDb struct {
	users map[string]string
}

func NewInMemoryUserDb() *InMemoryUserDb {
	return &InMemoryUserDb{
		users: make(map[string]string),
	}
}

func (userdb *InMemoryUserDb) MakeNewGuestName() string {
	id := 100 + rand.Intn(899)
	return fmt.Sprintf("Guest%v", id)
}

func (userdb *InMemoryUserDb) IsRegistered(name string) bool {
	_, ok := userdb.users[name]
	return ok
}

func (userdb *InMemoryUserDb) Register(name string, password string) error {
	if userdb.IsRegistered(name) {
		return errors.New(fmt.Sprintf("User %v is already registered", name))
	}
	userdb.users[name] = password
	return nil
}

func (userdb *InMemoryUserDb) Authenticate(name string, password string) error {
	if !userdb.IsRegistered(name) {
		return errors.New(fmt.Sprintf("User %v is not registered", name))
	} else if userdb.users[name] != password {
		return errors.New(fmt.Sprintf("Invalid password for user %v", name))
	}
	return nil
}
