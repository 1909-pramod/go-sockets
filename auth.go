package main

import "errors"

type User struct {
	Id string
}

type Auth struct {
	Users          []User
	UserWithToken  map[string]*User
	TokenAndUserId map[string]string
}

func (auth *Auth) userExists(userId string) bool {
	for _, user := range auth.Users {
		if user.Id == userId {
			return true
		}
	}
	return false
}

func (auth *Auth) getUserWithToken(token string) (*User, error) {
	user, exists := auth.UserWithToken[token]
	if exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}
