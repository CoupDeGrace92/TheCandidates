package main

import (
	"log"

	"github.com/CoupDeGrace92/candidates/internal/draft"
	"github.com/CoupDeGrace92/candidates/internal/game"
	"github.com/CoupDeGrace92/candidates/internal/scene"
	formats "github.com/CoupDeGrace92/candidates/internal/tournament/format"
	"github.com/hajimehoshi/ebiten/v2"
)

type MainApp struct {
	director *scene.Director
}

func (app *MainApp) Update() error {
	return app.director.Update()
}

func (app *MainApp) Draw(screen *ebiten.Image) {
	app.director.Draw(screen)
}

func (app *MainApp) Layout(outsideWidth, outsideHeight int) (int, int) {
	return app.director.Layout(outsideWidth, outsideHeight)
}

func main() {
	log.Printf("[System] Loading global chess sprite textures...")
	scene.LoadAssets("assets/images/GenericChessPiecesSprite.png")

	winCondition := formats.WinCondition{
		Type:  formats.ConditionFixedRounds,
		Limit: 10,
	}

	tourneyManager := formats.NewMatchTournament(winCondition, 3, 3, 3, 5, .2)

	log.Printf("[System] Seeding tournament participant profiles...")

	humanPlayer := game.NewDefaultProfile("player", true)
	botPlayer := game.NewDefaultProfile("bot", false)
	humanPlayer.Gold = 10
	botPlayer.Gold = 10
	botPlayer.BotTotalGold = 10

	roster := []*game.PlayerProfile{humanPlayer, botPlayer}
	tourneyManager.Initialize(roster)

	draftManager := draft.NewDraftManager(2)

	log.Printf("[System] Initializing central lifecycle Director State Machine")
	sceneDirector := scene.NewDirector(tourneyManager, roster, "player", draftManager)

	scene.RegenerateScaledUICache(60.0)
	app := &MainApp{
		director: sceneDirector,
	}

	ebiten.SetWindowSize(900, 750)
	ebiten.SetWindowTitle("Candidates: Auto-Chess")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	log.Printf("[System] Hardware graphics engine succesfully booted")
	if err := ebiten.RunGame(app); err != nil {
		log.Fatalf("Critical system error, game loop terminated unexpectedly: %v", err)
	}
}
