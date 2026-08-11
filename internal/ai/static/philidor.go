package static

import (
	"github.com/CoupDeGrace92/candidates/internal/ai"
	"github.com/CoupDeGrace92/candidates/internal/game"
)

var PhilidorCatalog = []ai.BotLayoutOptions{

	{
		Cost:       10,
		Color:      game.Black,
		BothColors: true,
		FEN:        "4k3/2pn1p2/3p1n2/4p3/8/8/8/8 w - - 0 1",
	},
}
