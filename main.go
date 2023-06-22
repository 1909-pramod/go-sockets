package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
)

var addr = flag.String("addr", ":8080", "http service address")

var Rooms = make(map[string]map[string]bool)
var hub = newHub()
var auth = &Auth{
	Users:          []User{},
	UserWithToken:  make(map[string]*User),
	TokenAndUserId: make(map[string]string),
}

var db *Db

func setupCorsResponse(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
}

func main() {
	flag.Parse()
	go hub.run()
	db = initDb()
	InitRandom()
	http.HandleFunc("/addUser", func(w http.ResponseWriter, r *http.Request) {
		setupCorsResponse(&w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Request not allowed", 405)
		}
		w.Header().Set("Content-Type", "application/json")
		decoder := json.NewDecoder(r.Body)
		var userData DbUser
		err := decoder.Decode(&userData)
		if err != nil {
			http.Error(w, "Invalid details", 400)
		}
		if db.UserExists(userData.Id) {
			http.Error(w, "User with user id: "+userData.Id+" already exists", 400)
		}
		db.AddUser(&userData)
		user := &User{
			Id: userData.Id,
		}
		token := GenerateToken()
		log.Printf("Created user with id %v", userData.Id)
		auth.UserWithToken[token] = user
		auth.TokenAndUserId[token] = user.Id
		json.NewEncoder(w).Encode(user)
	})

	http.HandleFunc("/getToken", func(w http.ResponseWriter, r *http.Request) {
		setupCorsResponse(&w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Request not allowed", 405)
		}
		w.Header().Set("Content-Type", "application/json")
		decoder := json.NewDecoder(r.Body)
		var userData DbUser
		err := decoder.Decode(&userData)
		if err != nil {
			http.Error(w, "Invalid details", 400)
			return
		}
		validUser, validErr := db.ValidateUser(&userData)
		if validErr != nil {
			http.Error(w, validErr.Error(), 400)
			return
		}
		tokenUser := &User{
			Id: validUser.Id,
		}
		tokenForUser, exists := auth.TokenAndUserId[tokenUser.Id]
		if exists {
			hub.UnregisterUser(tokenUser.Id)
			delete(auth.TokenAndUserId, tokenUser.Id)
			delete(auth.UserWithToken, tokenForUser)
		}
		token := GenerateToken()
		auth.TokenAndUserId[tokenUser.Id] = token
		auth.UserWithToken[token] = tokenUser
		mp := make(map[string]string)
		mp["id"] = tokenUser.Id
		mp["token"] = token
		log.Printf("token generated for user with id: %s, token: %s", tokenUser.Id, token)
		json.NewEncoder(w).Encode(mp)
	})
	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		setupCorsResponse(&w)
		token := strings.TrimPrefix(r.URL.Path, "/ws/")
		if auth.userExists(token) {
			http.Error(w, "User already connected", 400)
			return
		}
		user, err := auth.getUserWithToken(token)
		if err != nil {
			http.Error(w, "Invalid Token", 400)
			return
		}
		userId := user.Id
		log.Printf("Socket connected by user with id: %s \n", userId)
		serveWs(hub, w, r, userId)
	})
	log.Println("Server listening on port: 8080")
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("Error while running the server", err)
	}
}
