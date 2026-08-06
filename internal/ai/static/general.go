package static

import (
	"github.com/CoupDeGrace92/candidates/internal/ai"
	"github.com/CoupDeGrace92/candidates/internal/game"
)

//THIS FILE CONTAINS OPENINGS ELIGIBLE FOR SELECTION FOR ALL AIS

var GeneralCatalog = []ai.BotLayoutOptions{

	//=======================================
	//     STARTING POSITION ELIGIBLE
	//=======================================
	{
		Cost:       8,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/2B1P3/5N2/3P4/4K3 w - - 0 1",
	},

	{
		Cost:       8,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/8/8/5PPP/5RK1 w - - 0 1",
	},

	{
		Cost:       10,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/8/4B1P1/5PBP/6K1 w - - 0 1",
	},

	{
		Cost:       8,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/8/3P4/1BP1PP2/4K3 w - - 0 1",
	},

	{
		Cost:       8,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/8/6P1/4PPBP/6K1 w - - 0 1",
	},

	//==========================================
	//         10 - 20g
	//==========================================

	{
		Cost:       13,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/5NP1/5PBP/6K1 w - - 0 1",
	},

	{
		Cost:       17,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P1NP1/2P2PBP/6K1 w - - 0 1",
	},

	{
		Cost:       19,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P1NP1/2P2PBP/6K1 w - - 0 1",
	},

	{
		Cost:       17,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/P3P3/3P2P1/1PP2PBP/6K1 w - - 0 1",
	},

	//============================================
	//                21-30g
	//============================================

	{
		Cost:       23,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P1NP1/2P2PBP/5RK1 w - - 0 1",
	},

	{
		Cost:       21,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P1NP1/2PN1PBP/6K1 w - - 0 1",
	},

	{
		Cost:       26,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P1NP1/2PN1PBP/6K1 w - - 0 1",
	},

	{
		Cost:       26,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P2P1/2PN1PBP/4NRK1 w - - 0 1",
	},

	{
		Cost:       28,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P2P1/2PN1PBP/4NRK1 w - - 0 1",
	},

	{
		Cost:       28,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P2P1/2PN1PBP/2B1NRK1 w - - 0 1",
	},

	{
		Cost:       30,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4PP2/3P1NP1/2PN2BP/2B2RK1 w - - 0 1",
	},

	{
		Cost:       30,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4PP2/3P1NP1/2PN2RP/2B2BK1 w - - 0 1",
	},

	//============================================
	//           31-40g
	//============================================

	{
		Cost:       33,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/5P2/4P1P1/3P1N2/2PN2RP/2B2BK1 w - - 0 1",
	},

	{
		Cost:       33,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4PP2/3P1NP1/2P3RP/3Q1BK1 w - - 0 1",
	},

	{
		Cost:       32,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/P3PP2/3P1NP1/1PPN2BP/2B2RK1 w - - 0 1",
	},

	{
		Cost:       32,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/P3PP2/3P1NP1/2PN2BP/2B2RK1 w - - 0 1",
	},

	{
		Cost:       34,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/P3PP2/3P1NP1/2PN2BP/2B2RK1 w - - 0 1",
	},

	{
		Cost:       40,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4PP2/3P1NP1/2PN2BP/RRB3K1 w - - 0 1",
	},

	{
		Cost:       35,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4PP2/3P1NP1/1BPN2BP/R5K1 w - - 0 1",
	},

	{
		Cost:       35,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4PP2/3P1NP1/PPPN2BP/2B2RK1 w - - 0 1",
	},

	{
		Cost:       37,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4PP2/3P1NP1/2PN2BP/2BQ2K1 w - - 0 1",
	},

	{
		Cost:       33,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/8/3P1NP1/2PNPPBP/2BQ2K1 w - - 0 1",
	},

	{
		Cost:       36,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P1NP1/2PN1PBP/2BQ2K1 w - - 0 1",
	},

	//============================================
	//              41-50g
	//============================================

	{
		Cost:       41,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P1NP1/2PN1PBP/2BQ1RK1 w - - 0 1",
	},

	{
		Cost:       43,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P1NP1/PPPN1PBP/2BQ1RK1 w - - 0 1",
	},

	{
		Cost:       46,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/P3P3/3P1NP1/1PPN1PBP/2BQ1RK1 w - - 0 1",
	},

	{
		Cost:       48,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4P3/3P1NP1/2PN1PBP/R1BQ1RK1 w - - 0 1",
	},

	{
		Cost:       45,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/4PP2/3P1NP1/2PN2BP/2BQ1RK1 w - - 0 1",
	},

	{
		Cost:       49,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/5P2/4P1P1/3P1N2/2PN2RP/2BQ1BK1 w - - 0 1",
	},

	//============================================
	//          LATE GAME - 50+ gold
	//============================================

	{
		Cost:       50,
		Color:      game.White,
		BothColors: true,
		FEN:        "8/8/8/8/P3P3/2NP2P1/1PPN1PBP/R1BQ1RK1 w - - 0 1",
	},
}
