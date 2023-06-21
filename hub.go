package main

import (
	"bytes"
)

type Message struct {
	roomId []byte
	userId []byte
	data   []byte
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	users map[string]map[*Client]bool
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		users:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			user := h.users[client.userId]
			if user == nil {
				user = make(map[*Client]bool)
				h.users[client.userId] = user
			}
			user[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.users, client.userId)
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			if bytes.Compare(message.roomId, []byte("-")) != 0 {
				room, exists := Rooms[string(message.roomId)]
				if exists {
					for key, _ := range room {
						userId, exists := h.users[key]
						if exists {
							for client, _ := range userId {
								select {
								case client.send <- message:
								default:
									close(client.send)
									delete(h.clients, client)
								}
							}
						}
					}
				}
			}
			if bytes.Compare(message.userId, []byte("-")) != 0 {
				userId, exists := h.users[string(message.userId)]
				if exists {
					for client, _ := range userId {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(h.clients, client)
						}
					}
				}
			}
		}
	}
}
