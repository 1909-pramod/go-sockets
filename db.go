package main

import (
	"errors"
)

type DbUser struct {
	Id       string `json:id`
	Password string `json:password`
}

func (user *DbUser) CheckPassword(pswrd string) bool {
	return user.Password == pswrd
}

type Db struct {
	AllUsers map[string]*DbUser
}

func initDb() *Db {
	userMap := make(map[string]*DbUser)
	return &Db{
		AllUsers: userMap,
	}
}

func (db *Db) UserExists(Id string) bool {
	_, exists := db.AllUsers[Id]
	return exists
}

func (db *Db) ValidateUser(user *DbUser) (*DbUser, error) {
	usr, err := db.GetUserWithId(user.Id)
	if err != nil {
		return nil, err
	}
	if usr.Password != user.Password {
		return nil, errors.New("Invalid Password")
	}
	return usr, nil
}

func (db *Db) GetUserWithId(Id string) (*DbUser, error) {
	user, exists := db.AllUsers[Id]
	if exists {
		return user, nil
	}
	return nil, errors.New("User not found")
}

func (db *Db) AddUser(user *DbUser) {
	db.AllUsers[user.Id] = user
}
