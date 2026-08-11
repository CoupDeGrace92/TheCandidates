package scene

import (
	"log"

	"github.com/CoupDeGrace92/candidates/internal/ai"
	"github.com/CoupDeGrace92/candidates/internal/ai/static"
	"github.com/CoupDeGrace92/candidates/internal/draft"
	"github.com/CoupDeGrace92/candidates/internal/game"
	"github.com/CoupDeGrace92/candidates/internal/tournament"
	"github.com/hajimehoshi/ebiten/v2"
)

type Scene interface {
	Update() error
	Draw(screen *ebiten.Image)
	Layout(outsideWidth, outsideHeight int) (int, int)
}

type Director struct {
	currentScene Scene
	tourney      tournament.TournamentManager
	players      []*game.PlayerProfile
	draftManager *draft.DraftManager

	localClientID string

	//Persistant window dimensions
	lastWinW int
	lastWinH int
}

func NewDirector(tourney tournament.TournamentManager, participants []*game.PlayerProfile, localID string, dm *draft.DraftManager) *Director {
	d := &Director{
		tourney:       tourney,
		players:       participants,
		localClientID: localID,
		draftManager:  dm,
		lastWinW:      800,
		lastWinH:      800,
	}

	// Boot straight into Round 1 drafting phase setup
	d.EnterDraftPhase()
	return d
}

func (d *Director) EnterDraftPhase() {
	if d.tourney.IsCompleted() {
		log.Printf("Tournament concluded! Moving to final leaderboard cross-tables view dashboard.")
		return
	}

	_, err := d.tourney.GeneratePairings()
	if err != nil {
		log.Fatalf("Director Failure: Pairing matrix calculation aborted: %v", err)
	}

	var localProfile *game.PlayerProfile
	var localAssignedColor game.PieceColor

	activePairings := d.getCurrentTourneyPairings()

	for _, pair := range activePairings {
		if pair.WhitePlayer.PlayerID == d.localClientID {
			localProfile = pair.WhitePlayer
			localAssignedColor = game.White
			break
		} else if pair.BlackPlayer.PlayerID == d.localClientID {
			localProfile = pair.BlackPlayer
			localAssignedColor = game.Black
			break
		}
	}
	localProfile.SwitchColorIfDifferent(localAssignedColor)
	d.currentScene = NewShopScene(localProfile, d.draftManager, d)
}

func (d *Director) getCurrentTourneyPairings() []tournament.Pairing {
	pairings, _ := d.tourney.GeneratePairings()
	return pairings
}

func (d *Director) CompleteDraftAndEnterBattle() {
	pairings, err := d.tourney.GeneratePairings()
	if err != nil || len(pairings) == 0 {
		log.Fatalf("Director Transition Error: Pairing retrieval failed: %v", err)
	}
	activePair := pairings[0]

	if !activePair.WhitePlayer.IsHuman {
		ai.SelectAndDeployStaticLayout(activePair.WhitePlayer, static.GeneralCatalog, game.White)
	}
	if !activePair.BlackPlayer.IsHuman {
		ai.SelectAndDeployStaticLayout(activePair.BlackPlayer, static.GeneralCatalog, game.Black)
	}

	globalBoard, _, err := game.MergePlayerPlacements(activePair.WhitePlayer, activePair.BlackPlayer)
	if err != nil {
		log.Printf("Merge Phase Interception Warning: %v", err)
	}

	log.Printf("Draft configurations unified! Initializing MatchState parameters...")

	matchState := &game.MatchState{
		Board:           globalBoard,
		WhitePlayer:     activePair.WhitePlayer,
		BlackPlayer:     activePair.BlackPlayer,
		ActiveColor:     game.White,
		HalfMoveClock:   0,
		FullMoveNumber:  1,
		CastlingRights:  "-",
		EnPassantTarget: "-",
	}

	binPath := "assets/engines/stockfish"
	battleScene, err := NewBattleScene(binPath, matchState, d)
	if err != nil {
		log.Printf("Fatal Director Transition Error: Failed to spawn Stockfish BattleScene: %v", err)
		return
	}

	d.currentScene = battleScene
}

func (d *Director) ConcludeBattleAndReturnToShop(outcome tournament.MatchOutcome) {
	d.tourney.ResolveGameOutcome(outcome)
	d.EnterDraftPhase()
}

func (d *Director) Update() error             { return d.currentScene.Update() }
func (d *Director) Draw(screen *ebiten.Image) { d.currentScene.Draw(screen) }
func (d *Director) Layout(w, h int) (int, int) {
	d.lastWinW = w
	d.lastWinH = h
	return d.currentScene.Layout(w, h)
}
