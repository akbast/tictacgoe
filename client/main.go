package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"mitarbeiterprojekt/tictactoe/shared"
	"net/url"
)

var commandChannel = make(chan shared.Command)

var addr = flag.String("addr", "localhost:8080", "http service address")
var playerId int
var sessionId string

func main() {

	flag.Parse()
	log.SetFlags(0)

	url := url.URL{Scheme: "ws", Host: *addr, Path: "/game"}
	log.Printf("connecting to %s", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	done := make(chan struct{})
	go readFromSocket(done, conn)()

	go addAsNewPlayer()

	for {
		select {
		case <-done:
			return
		case command := <-commandChannel:
			if sendToSocket(conn, command) {
				return
			}
		}
	}
}

func sendToSocket(conn *websocket.Conn, command shared.Command) bool {
	err := conn.WriteJSON(command)
	if err != nil {
		log.Println("write:", err)
		return true
	}
	return false
}

func addAsNewPlayer() {
	commandChannel <- shared.Command{Name: shared.ServerCommandAddNewPlayer, Params: nil}
}

func readFromSocket(done chan struct{}, conn *websocket.Conn) func() {
	return func() {
		defer close(done)
		for {
			var command shared.Command
			err := conn.ReadJSON(&command)
			if err != nil {
				return
			}
			evaluateMessage(command, conn)

		}
	}
}

func evaluateMessage(command shared.Command, conn *websocket.Conn) {
	switch command.Name {
	case shared.ClientCommandPlayerAdded:
		onPlayerAdded(command)
	case shared.ClientCommandGameBegins:
		onGameBegins(command)
	case shared.ClientCommandDisplayBoard:
		onDisplayBoard(command)
	case shared.ClientCommandAskForPlay:
		onAskForPlay(command)
	case shared.ClientWaitForMove:
		onWaitForMove(command)
	case shared.ClientWrongMove:
		onWrongMove(command)
	case shared.ClientGameEnds:
		onGameEnds(command)
	}
}

func onGameEnds(command shared.Command) {
	info := command.Params["info"]
	log.Println(info)
}

func onWrongMove(command shared.Command) {
	info := command.Params["info"]
	log.Println(info)
}

func onWaitForMove(command shared.Command) {
	log.Println("Wait for other players move.")
}

func onAskForPlay(command shared.Command) {
	log.Println("Make your move")
	var moveString string
	fmt.Scan(&moveString)
	// fmt.Println("moveInt is", moveInt)
	params := make(map[string]interface{})
	params["id"] = playerId
	params["move"] = moveString
	params["sessionId"] = sessionId
	commandChannel <- shared.Command{Name: shared.ServerCommandUserMove, Params: params}
}

func onDisplayBoard(command shared.Command) {
	log.Println("-Actual status of board-")

	var boardFields = (command.Params["boardFields"]).([]interface{})
	printBoard(boardFields)
}

func onGameBegins(command shared.Command) {
	log.Println("Game is beginning. Fasten your seatbelts.")
}

func printBoard(boardFields []interface{}) {
	var boardFieldsNew = [9]string{}
	for i, v := range boardFields {
		boardFieldsNew[i] = v.(string)
	}
	board := shared.Board{Fields: boardFieldsNew}
	log.Println(board)
}

func onPlayerAdded(command shared.Command) {
	id := int(command.Params["playerId"].(float64))
	sessionId = command.Params["sessionId"].(string)
	playerId = id
	log.Printf("Your player Id is %d\n", playerId)
}
