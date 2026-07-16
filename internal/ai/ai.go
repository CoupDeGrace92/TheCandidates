package ai

import (
	"math/rand"
	"time"

	"github.com/CoupDeGrace92/candidates/internal/game"
)

type StaticLayout struct {
	Board game.BoardState
}

type BotLayoutOptions struct {
	Cost       int
	Color      game.PieceColor
	BothColors bool
	FEN        string
}

func SelectAndDeployStaticLayout(profile *game.PlayerProfile, catalog []BotLayoutOptions, botColor game.PieceColor) StaticLayout {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	currentGold := profile.BotTotalGold

	var eligibleChoices []BotLayoutOptions
	minGoldFloor := currentGold - 5
	if minGoldFloor < 0 {
		minGoldFloor = 0
	}

	for _, option := range catalog {
		if option.Cost >= minGoldFloor && option.Cost <= currentGold && (option.Color == profile.Color || option.BothColors) {
			eligibleChoices = append(eligibleChoices, option)
		}
	}

	if len(eligibleChoices) == 0 {
		for _, option := range catalog {
			if option.Cost <= currentGold {
				eligibleChoices = append(eligibleChoices, option)
			}
		}
	}

	chosenOption := eligibleChoices[rng.Intn(len(eligibleChoices))]

	profile.Gold -= profile.BotTotalGold - chosenOption.Cost
	extractedBoard := game.ParseFENForColor(chosenOption.FEN, game.White)

	board := make(game.BoardState)
	for loc, piece := range extractedBoard {
		piece.Color = botColor
		board[loc] = piece
	}

	bb := profile.BoardAndBench
	*bb.Board = board
	bb.Bench = []game.Piece{} // Static bots maintain an empty bench frame environment

	return StaticLayout{Board: board}
}
