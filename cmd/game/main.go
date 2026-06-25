package main

import (
	"log"

	"github.com/CoupDeGrace92/candidates/internal/draft"
	"github.com/CoupDeGrace92/candidates/internal/game"
	"github.com/CoupDeGrace92/candidates/internal/scene"
	"github.com/hajimehoshi/ebiten/v2"
)

type MainGameApp struct {
	currentScene *scene.ShopScene
}

func (m *MainGameApp) Update() error {
	return m.currentScene.Update()
}

func (m *MainGameApp) Draw(screen *ebiten.Image) {
	m.currentScene.Draw(screen)
}

func (m *MainGameApp) Layout(outsideWidth, outsideHeigh int) (int, int) {
	return m.currentScene.Layout(outsideWidth, outsideHeigh)
}

//Test for for the resolution phase
/*func main() {
	scene.LoadAssets("assets/images/GenericChessPiecesSprite.png")

	wbs := make(game.BoardState)
	bbs := make(game.BoardState)

	wbs[game.Location{File: 5, Rank: 2}] = game.Piece{Type: game.Pawn, Color: game.White}
	wbs[game.Location{File: 5, Rank: 1}] = game.Piece{Type: game.King, Color: game.White}
	bbs[game.Location{File: 5, Rank: 8}] = game.Piece{Type: game.King, Color: game.Black}

	//HardWired white and black player defaults:
	WhitePieces := game.PlayerPieces{Board: &wbs}
	BlackPieces := game.PlayerPieces{Board: &bbs}

	player1 := game.NewDefaultProfile("player1", true)
	player1.BoardAndBench = &WhitePieces

	player2 := game.NewDefaultProfile("player2", true)
	player2.BoardAndBench = &BlackPieces

	initialMatch := &game.MatchState{
		ActiveColor:     game.White,
		CastlingRights:  "-",
		EnPassantTarget: "-",
		HalfMoveClock:   0,
		FullMoveNumber:  1,
		WhitePlayer:     player1,
		BlackPlayer:     player2,
	}

	battle, err := scene.NewBattleScene("assets/engines/stockfish", initialMatch)
	if err != nil {
		log.Fatalf("Fatal: failed to open visualization environment")
	}

	defer battle.Destroy()

	app := &MainGameApp{currentScene: battle}

	ebiten.SetWindowSize(640, 640)
	ebiten.SetWindowTitle("The Candidates - Engine Combat Tester")

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
} */

func main() {
	profile := game.NewDefaultProfile("white", true)
	manager := draft.NewDraftManager(10)
	shop := scene.NewShopScene(profile, manager)
	defer shop.Destroy()
	app := &MainGameApp{currentScene: shop}

	ebiten.SetWindowSize(640, 640)
	ebiten.SetWindowTitle("The Candidates - Scaffolding Draft Shop")

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
