package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
)

type ConnectionRequest struct {
	Token  string `json:token`
	UserId string `json:userId`
}

type AcceptConnectionBody struct {
	UserToken       string `json:userToken`
	ConnectionToken string `json:connectionToken`
}

var addr = flag.String("addr", ":8080", "http service address")

var Rooms = make(map[string]map[string]bool)
var hub = newHub()
var auth = &Auth{
	Users:          []User{},
	UserWithToken:  make(map[string]*User),
	TokenAndUserId: make(map[string]string),
}

var db *Db
var connectionRequest *UserRequests

func setupCorsResponse(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
}

func main() {
	flag.Parse()
	go hub.run()
	db = initDb()
	connectionRequest = InitUserRequests()
	InitRandom()

	http.HandleFunc("/addRequest", func(w http.ResponseWriter, r *http.Request) {
		setupCorsResponse(&w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Request not allowed", 405)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		decoder := json.NewDecoder(r.Body)
		var connectData ConnectionRequest
		err := decoder.Decode(&connectData)
		if err != nil {
			http.Error(w, "Invalid details", 400)
		}
		user, err := auth.getUserWithToken(connectData.Token)
		if err != nil {
			http.Error(w, "Invalid Token", 400)
			return
		}
		if !auth.userExists(user.Id) {
			http.Error(w, "Please login to access this", 400)
			return
		}
		exists := connectionRequest.CheckConnectionExists(user.Id, connectData.UserId)
		if exists {
			http.Error(w, "Already connected", 400)
			return
		}
		connectionRequest.AddRequests(user.Id, connectData.UserId)
		hubUser, userConnected := hub.users[connectData.UserId]
		if !userConnected {
			http.Error(w, "User not connected", 400)
			return
		}
		hubUser.send <- Message{
			data:        []byte(connectData.Token),
			roomId:      []byte("-"),
			userId:      []byte(connectData.UserId),
			messageType: []byte("CONNECTION_REQUEST"),
		}
		resp := make(map[string]string)
		resp["message"] = "Requested"
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/acceptRequest", func(w http.ResponseWriter, r *http.Request) {
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
		var connectionData AcceptConnectionBody
		err := decoder.Decode(&connectionData)
		if err != nil {
			http.Error(w, "Invalid details", 400)
		}
		user, err := auth.getUserWithToken(connectionData.UserToken)
		if err != nil {
			http.Error(w, "Invalid Token", 400)
			return
		}
		if !auth.userExists(user.Id) {
			http.Error(w, "Please login to access this", 400)
			return
		}
		log.Printf("connection data %v\n", connectionData)
		connUser, connErr := auth.getUserWithToken(connectionData.ConnectionToken)
		if connErr != nil {
			http.Error(w, "Invalid Token", 400)
			return
		}
		if !auth.userExists(connUser.Id) {
			http.Error(w, "User disconnected", 400)
			return
		}
		if !connectionRequest.CheckConnectionExists(user.Id, connUser.Id) {
			connectionRequest.CreateConnections(user.Id, connUser.Id)
		}
		hubUser, userConnected := hub.users[user.Id]
		if !userConnected {
			http.Error(w, "User not connected", 400)
			return
		}
		hubUser.send <- Message{
			data:        []byte(connectionData.ConnectionToken),
			roomId:      []byte("-"),
			userId:      []byte(connUser.Id),
			messageType: []byte("CONNECTION_REQUEST_ACCEPTED"),
		}
		hubUser1, userConnected1 := hub.users[connUser.Id]
		if !userConnected1 {
			http.Error(w, "User not connected", 400)
			return
		}
		hubUser1.send <- Message{
			data:        []byte(connectionData.UserToken),
			roomId:      []byte("-"),
			userId:      []byte(user.Id),
			messageType: []byte("CONNECTION_REQUEST_ACCEPTED"),
		}
		log.Printf("connection created %V", connectionRequest.UserConnections)
		resp := make(map[string]string)
		resp["message"] = "Requested"
		json.NewEncoder(w).Encode(resp)
	})

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
