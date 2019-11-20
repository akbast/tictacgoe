package shared

const (
	SignPlayerOne             = " X "
	SignPlayerTwo             = " O "
	ServerCommandAddNewPlayer = "ADD_NEW_PLAYER"
	ClientCommandPlayerAdded  = "PLAYER_ADDED"
	ClientCommandGameBegins   = "GAME_BEGINS"
	ClientCommandDisplayBoard = "DISPLAY_BOARD"
	ClientCommandAskForPlay   = "ASK_FOR_PLAY"
	ServerCommandUserMove     = "USER_MADE_MOVE"
	ClientWaitForMove         = "WAIT_FOR_MOVE"
	ClientWrongMove           = "WRONG_MOVE"
	ClientGameEnds            = "GAME_ENDS"
)

type Command struct {
	Name   string
	Params map[string]interface{}
}

func InitializeBoard() Board {
	return Board{0, [9]string{
		"   ",
		"   ",
		"   ",
		"   ",
		"   ",
		"   ",
		"   ",
		"   ",
		"   ",
	}}
}
