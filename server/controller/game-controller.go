package controller

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"mitarbeiterprojekt/tictactoe/server/model"
	"mitarbeiterprojekt/tictactoe/shared"
	"strconv"
)

type GameController struct {
	SessionId string
	Players   map[int]model.Player
	board     shared.Board
}

func (controller GameController) Initialize(sessionId string) *GameController {
	return &GameController{
		SessionId: sessionId,
		Players:   make(map[int]model.Player),
		board:     shared.Board{}.InitializeBoard(),
	}
}

func (controller *GameController) CreateNewPlayer(conn *websocket.Conn) {
	playerId := len(controller.Players)
	sign := " X "
	if playerId == 1 {
		sign = " O "
	}
	controller.Players[len(controller.Players)] = model.Player{
		Id:         playerId,
		Sign:       sign,
		Connection: conn,
	}

	controller.informPlayer(playerId)
}

func (controller GameController) informPlayer(playerId int) {
	params := make(map[string]interface{})
	params["sessionId"] = controller.SessionId
	params["playerId"] = playerId
	command := shared.Command{Name: shared.ClientCommandPlayerAdded, Params: params}

	player := controller.Players[playerId]
	player.SendCommand(command)

	controller.informPlayersForGameBegin()
}

func (controller GameController) informPlayersForGameBegin() {
	gameBeginsCommand := shared.Command{Name: shared.ClientCommandGameBegins, Params: nil}

	if len(controller.Players) < 2 {
		return
	}

	for _, player := range controller.Players {
		player.SendCommand(gameBeginsCommand)
	}

	controller.decideTurnsAndAskForMovement()
}

func (controller GameController) prepareDisplayBoardCommand() shared.Command {
	params := make(map[string]interface{})
	params["boardFields"] = controller.board.Fields
	displayBoardCommand := shared.Command{Name: shared.ClientCommandDisplayBoard, Params: params}
	return displayBoardCommand
}

func (controller GameController) decideTurnsAndAskForMovement() {
	winnerId := controller.checkForWin()
	if winnerId != -1 {
		controller.informPlayersForGameEnd(winnerId)
		return
	}

	playerId, waiterId := 0, 0
	if controller.board.TurnNumber%2 == 1 {
		fmt.Println("Player 0's turn")
		playerId, waiterId = 0, 1
	} else {
		fmt.Println("Player 1's turn")
		playerId, waiterId = 1, 0
	}
	controller.askForPlay(playerId, waiterId)
}

func prepareGameEndCommand(info string) shared.Command {
	params := make(map[string]interface{})
	params["info"] = info
	return shared.Command{Name: shared.ClientGameEnds, Params: params}
}

func (controller GameController) checkForWin() int {
	numberArray := make([]int, 9)
	for i, v := range controller.board.Fields {
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

func (controller GameController) askForPlay(playerId, waiterId int) {
	displayBoardCommand := controller.prepareDisplayBoardCommand()
	playCommand := shared.Command{Name: shared.ClientCommandAskForPlay, Params: nil}
	player := controller.Players[playerId]
	player.SendCommand(displayBoardCommand)
	player.SendCommand(playCommand)

	waitCommand := shared.Command{Name: shared.ClientWaitForMove, Params: nil}
	waiter := controller.Players[waiterId]
	waiter.SendCommand(waitCommand)
}

func (controller GameController) informPlayerForWrongMovement(player model.Player, move int) {
	params := make(map[string]interface{})
	params["info"] = fmt.Sprintf("Wrong move %d. It is blocked already.", move)
	wrongMoveCommand := shared.Command{Name: shared.ClientWrongMove, Params: params}
	player.SendCommand(wrongMoveCommand)

	controller.decideTurnsAndAskForMovement()
}

func (controller *GameController) EvaluateMovement(command shared.Command) {
	log.Println("Evaluate movement", command)
	playerId := int(command.Params["id"].(float64))
	moveString := command.Params["move"].(string)
	move, _ := strconv.Atoi(moveString)

	player, ok := controller.Players[playerId]
	if !ok {
		log.Printf("Player could not be found for Id %d \n", playerId)
	}

	log.Printf("Making move for player %d to %d\n", playerId, move)

	movePlace := &controller.board.Fields[move]
	if *movePlace == "   " {
		log.Println("Move can be placed.")
		*movePlace = player.Sign
		controller.board.TurnNumber = controller.board.TurnNumber + 1
		log.Println(controller.board.TurnNumber)
		controller.decideTurnsAndAskForMovement()
	} else {
		log.Printf("Move cannot be placed. It is already marked for %d\n", player.Id)
		controller.informPlayerForWrongMovement(player, move)
	}
}

func (controller GameController) informPlayersForGameEnd(winnerId int) {
	winner := controller.Players[winnerId]
	var loser model.Player
	for _, player := range controller.Players {
		if player.Id != winnerId {
			loser = player
		}
	}

	winner.SendCommand(prepareGameEndCommand("Congratulations! You won!"))
	loser.SendCommand(prepareGameEndCommand("You Lose!"))
	controller.resetState()
}

func (controller GameController) resetState() {
	// close connections
	for _, player := range controller.Players {
		player.CloseConnection()
	}
	// remove players
	controller.Players = make(map[int]model.Player)

	// reset board
	controller.board = shared.Board{}.InitializeBoard()
}
