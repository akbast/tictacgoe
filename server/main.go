package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"mitarbeiterprojekt/tictactoe/server/model"
	"mitarbeiterprojekt/tictactoe/shared"
	"net/http"
	"strconv"
)

var players = make(map[int]model.Player)
var board = shared.InitializeBoard()

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	http.HandleFunc("/game", game)

	log.Fatal(http.ListenAndServe(":8080", nil))
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
		addNewPlayer(conn)
	case shared.ServerCommandUserMove:
		evaluateMovement(command)
	default:
		log.Println("Unknown command", command.Name)
	}
}

func addNewPlayer(conn *websocket.Conn) {
	playerId := len(players)
	sign := " X "
	if playerId == 1 {
		sign = " O "
	}
	players[len(players)] = model.Player{playerId, sign, conn}
	log.Printf("New Player with Id %d added\n", playerId)
	informPlayer(playerId, conn)
}

func informPlayer(playerId int, conn *websocket.Conn) {
	params := make(map[string]interface{})
	params["id"] = playerId
	command := shared.Command{Name: shared.ClientCommandPlayerAdded, Params: params}

	err := conn.WriteJSON(command)
	if err != nil {
		fmt.Println("error on writing: ", err)
	}

	informPlayersForGameBegin()
}

func informPlayersForGameBegin() {
	gameBeginsCommand := shared.Command{Name: shared.ClientCommandGameBegins, Params: nil}

	if len(players) < 2 {
		return
	}

	for _, player := range players {
		player.SendCommand(gameBeginsCommand)
	}

	decideTurnsAndAskForMovement()
}

func prepareDisplayBoardCommand() shared.Command {
	params := make(map[string]interface{})
	params["boardFields"] = board.Fields
	displayBoardCommand := shared.Command{Name: shared.ClientCommandDisplayBoard, Params: params}
	return displayBoardCommand
}

func decideTurnsAndAskForMovement() {
	winnerId := checkForWin()
	if winnerId != -1 {
		informPlayersForGameEnd(winnerId)
		return
	}

	playerId, waiterId := 0, 0
	if board.TurnNumber%2 == 1 {
		fmt.Println("Player 0's turn")
		playerId, waiterId = 0, 1
	} else {
		fmt.Println("Player 1's turn")
		playerId, waiterId = 1, 0
	}
	askForPlay(playerId, waiterId)

}

func informPlayersForGameEnd(winnerId int) {
	winner := players[winnerId]
	var loser model.Player
	for _, player := range players {
		if player.Id != winnerId {
			loser = player
		}
	}

	winner.SendCommand(prepareGameEndCommand("Congratulations! You won!"))
	loser.SendCommand(prepareGameEndCommand("You Lose!"))
	resetState()
}

func resetState() {
	// close connections
	for _, player := range players {
		player.CloseConnection()
	}
	// remove players
	players = make(map[int]model.Player)

	// reset board
	board = shared.InitializeBoard()
}

func prepareGameEndCommand(info string) shared.Command {
	params := make(map[string]interface{})
	params["info"] = info
	return shared.Command{Name: shared.ClientGameEnds, Params: params}
}

func checkForWin() int {
	numberArray := make([]int, 9)
	for i, v := range board.Fields {
		value := 0
		if v == shared.SignPlayerOne {
			value = 1
		} else if v == shared.SignPlayerTwo {
			value = 10
		}
		numberArray[i] = value
	}

	resultArray := make([]int, 9)
	resultArray[0] = numberArray[0] + numberArray[1] + numberArray[2]
	resultArray[1] = numberArray[3] + numberArray[4] + numberArray[5]
	resultArray[2] = numberArray[6] + numberArray[7] + numberArray[8]
	resultArray[3] = numberArray[0] + numberArray[3] + numberArray[6]
	resultArray[4] = numberArray[1] + numberArray[4] + numberArray[7]
	resultArray[5] = numberArray[2] + numberArray[5] + numberArray[8]
	resultArray[6] = numberArray[0] + numberArray[4] + numberArray[8]
	resultArray[7] = numberArray[2] + numberArray[4] + numberArray[6]

	for _, result := range resultArray {
		if result == 3 {
			// Player One Wins
			return 0
		} else if result == 30 {
			// Player Two Wins
			return 1
		}
	}
	return -1
}

func askForPlay(playerId, waiterId int) {
	displayBoardCommand := prepareDisplayBoardCommand()
	playCommand := shared.Command{Name: shared.ClientCommandAskForPlay, Params: nil}
	player := players[playerId]
	player.SendCommand(displayBoardCommand)
	player.SendCommand(playCommand)

	waitCommand := shared.Command{Name: shared.ClientWaitForMove, Params: nil}
	waiter := players[waiterId]
	waiter.SendCommand(waitCommand)
}

func informPlayerForWrongMovement(player model.Player, move int) {
	params := make(map[string]interface{})
	params["info"] = fmt.Sprintf("Wrong move %d. It is blocked already.", move)
	wrongMoveCommand := shared.Command{Name: shared.ClientWrongMove, Params: params}
	player.SendCommand(wrongMoveCommand)

	decideTurnsAndAskForMovement()
}

func evaluateMovement(command shared.Command) {
	log.Println("Evaluate movement", command)
	playerId := int(command.Params["id"].(float64))
	moveString := command.Params["move"].(string)
	move, _ := strconv.Atoi(moveString)

	player, ok := players[playerId]
	if !ok {
		log.Printf("Player could not be found for Id %d \n", playerId)
	}

	log.Printf("Making move for player %d to %d\n", playerId, move)

	movePlace := &board.Fields[move]
	if *movePlace == "   " {
		log.Println("Move can be placed.")
		*movePlace = player.Sign
		board.TurnNumber = board.TurnNumber + 1
		decideTurnsAndAskForMovement()
	} else {
		log.Printf("Move cannot be placed. It is already marked for %d\n", player.Id)
		informPlayerForWrongMovement(player, move)
	}
}
