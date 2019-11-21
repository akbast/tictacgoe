package model

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"tictacgoe/shared"
)

type Player struct {
	Id         int
	Sign       string
	Connection *websocket.Conn
}

func (player Player) SendCommand(command shared.Command) {
	err := player.Connection.WriteJSON(command)
	if err != nil {
		fmt.Println("error on writing: ", err)
	}
}

func (player Player) CloseConnection() {
	err := player.Connection.Close()
	if err != nil {
		fmt.Println("Error on closing connection", err)
	}
	log.Printf("Connection to %d closed.", player.Id)
}
