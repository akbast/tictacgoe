package main

import "fmt"

const (
	Player1   = " X "
	Player2   = " O "
	NotUsed      = "   "
)

type Board struct {
	Fields [9]string
}

func (board Board) String() string {
	boardString := ""
	for i, v := range board.Fields {
		if (i%3) == 0 {
			boardString = boardString + "\n"

		} else {
			boardString = boardString + "|"
		}
		boardString = boardString + v
	}
	return boardString
}

func main() {
	board := Board{[9]string{Player1, Player2, NotUsed, Player1, NotUsed, NotUsed, NotUsed, Player2, Player1}}

	fmt.Println(board)
}