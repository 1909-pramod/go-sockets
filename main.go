package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
)

var addr = flag.String("addr", ":8080", "http service address")

var Rooms = make(map[string]map[string]bool)

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()
	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		userId := strings.TrimPrefix(r.URL.Path, "/ws/")
		log.Printf("Socket connected by user with id: %s \n", userId)
		serveWs(hub, w, r, userId)
	})
	log.Println("Server listening on port: 8080")
	err  := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("Error while running the server", err)
	}
}
