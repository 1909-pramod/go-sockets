package main

import (
	"math/rand"
	"time"
)

const TOKEN_SIZE = 10

func InitRandom() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func getRandomToken(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GenerateToken() string {
	token := getRandomToken(TOKEN_SIZE)
	user, _ := auth.getUserWithToken(token)
	if user != nil {
		return GenerateToken()
	}
	return token
}
