package tools

import "context"

func HandleMakeMove(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	board, _ :=getString(args, "board")
	pos, _ :=getInt(args, "position")
	player, _ :=getString(args, "player")
	if len(board) != 9 {
		return err("invalid board length")
}

	if pos < 0 || pos > 8 {
		return err("invalid position")
}

	if player != "X" && player != "O" {
		return err("invalid player")
}

	if board[pos] != ' ' {
		return err("position already taken")
}

	newBoard := []byte(board)
	newBoard[pos] = player[0]
	return ok(string(newBoard))
}

func HandleNewGame(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("         ")
}