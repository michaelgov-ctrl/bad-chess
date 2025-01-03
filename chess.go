package main

/*
import (
	"fmt"
	"os"
	"os/exec"

	"github.com/notnil/chess"
)

type PlayerColor int

const (
	Light PlayerColor = iota
	Dark
	NoColor
)

type ChessGame struct {
	Game *chess.Game
	Turn PlayerColor
}

func newChessGame() *ChessGame {
	return &ChessGame{
		Game: chess.NewGame(),
		Turn: Light,
	}
}

func (cg *ChessGame) playTillOver() {
	for cg.Game.Outcome() == chess.NoOutcome {
		move := cg.getTurnInput()

		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()

		if err := cg.Game.MoveStr(move); err != nil {
			fmt.Println("invalid move")
			continue
		}

		cg.changeTurn()
		fmt.Println(cg.Game.Position().Board().Draw())
	}
}

func (cg *ChessGame) getTurnInput() string {
	fmt.Printf("%s's turn to move\n", cg.Turn)
	var input string
	fmt.Scanln(&input)
	return input
}

func (cg *ChessGame) changeTurn() {
	if cg.Turn == Light {
		cg.Turn = Dark
	} else {
		cg.Turn = Light
	}
}

func (pc PlayerColor) String() string {
	if pc == Light {
		return "light"
	}

	return "dark"
}
*/
