package shared

type Board struct {
	TurnNumber int
	Fields     [9]string
}

func (board Board) String() string {
	boardString := ""
	for i, v := range board.Fields {
		if (i % 3) == 0 {
			boardString = boardString + "\n"

		} else {
			boardString = boardString + "|"
		}
		boardString = boardString + v
	}
	return boardString
}

func (board *Board) GetListOfValues() []string {
	valueList := make([]string, 9)
	for _, value := range board.Fields {
		valueList = append(valueList, value)
	}
	return valueList
}
