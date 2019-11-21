package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"mitarbeiterprojekt/tictactoe/server/controller"
	"mitarbeiterprojekt/tictactoe/shared"
	"net/http"
)

var gameControllers = make(map[string]*controller.GameController)

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	http.HandleFunc("/game", game)

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func game(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer conn.Close()
	for {
		var command shared.Command
		// Read in a new message as JSON and map it to a Message object
		err := conn.ReadJSON(&command)
		if err != nil {
			log.Println("read:", err)
			break
		}
		evaluateMessage(command, conn)
	}
}

func evaluateMessage(command shared.Command, conn *websocket.Conn) {
	log.Printf("recv: %s", command)

	switch command.Name {
	case shared.ServerCommandAddNewPlayer:
		addNewPlayerNew(conn)
	case shared.ServerCommandUserMove:
		evaluateMovement(command, conn)
	default:
		log.Println("Unknown command", command.Name)
	}
}

func evaluateMovement(command shared.Command, conn *websocket.Conn) {
	sessionId := command.Params["sessionId"].(string)
	gameController, ok := gameControllers[sessionId]
	if !ok {
		log.Printf("Cannot find any session for %s", sessionId)
		conn.Close()
	}
	gameController.EvaluateMovement(command)
}

func addNewPlayerNew(conn *websocket.Conn) {
	var gameController controller.GameController
	for _, gCon := range gameControllers {
		if len(gCon.Players) == 1 {
			gCon.CreateNewPlayer(conn)
			return
		}
	}
	sessionId := uuid.New().String()
	gameController = *controller.GameController{}.Initialize(sessionId)
	gameControllers[sessionId] = &gameController
	gameController.CreateNewPlayer(conn)
}
