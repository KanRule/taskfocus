// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"strings"
)

//Hub maintains the set of active clients and broadcasts messages to the clients.
type MsgCarrier struct {
	msg          []byte
	originClient *Client
}
type Hub struct {
	// Registered clients.
	clients map[*Client]string

	// Inbound messages from the clients.
	broadcast chan MsgCarrier

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan MsgCarrier),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]string),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = client.wallID
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case msgData := <-h.broadcast:
			msg := msgData.msg
			msgsplit := strings.Split(string(msg), " ~ ~ ")
			if len(msgsplit) != 2 {
				log.Println("Socket Error: Poorly Formatted Message")
			}
			//Useful for socket debugging
			log.Println("Socket msgtype:", msgsplit[0])
			log.Println("Socket msg:", msgsplit[1])
			switch msgsplit[0] {
			case "init":
				socketInitWall(msgsplit[1], msgData.originClient)
			case "addCard":
				socketAddCard(msgsplit[1], h)
			case "deleteCard":
				socketDeleteCard(msgsplit[1], h)
			case "addBoard":
				socketAddBoard(msgsplit[1], h)
			case "addWall":
				socketAddWall(msgsplit[1], msgData.originClient)
			case "changeWallName":
				socketChangeWallName(msgsplit[1], h)
			case "changeBoardName":
				socketChangeBoardName(msgsplit[1], h)
			case "changeCardTitle":
				socketChangeCardTitle(msgsplit[1], h)
			case "changeCardDetails":
				socketChangeCardDetails(msgsplit[1], h)
			case "moveCard":
				socketMoveCard(msgsplit[1], h)
			case "addChecklistItem":
				socketAddChecklistItem(msgsplit[1], h)
			case "updateChecklistItem":
				socketUpdateChecklistItem(msgsplit[1], h)
			case "deleteChecklistItem":
				socketDeleteChecklistItem(msgsplit[1], h)
			}
		}
	}
}
func (h *Hub) broadcastAll(msg []byte) {
	for client := range h.clients {
		select {
		case client.send <- msg:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) broadcastChannel(wallID string, msg []byte) {
	for client := range h.clients {
		if client.wallID == wallID {
			select {
			case client.send <- msg:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}
