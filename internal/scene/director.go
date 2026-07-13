package scene

import (
	"log"

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
	d.currentScene = NewShopScene(localProfile, d.draftManager)
}

func (d *Director) getCurrentTourneyPairings() []tournament.Pairing {
	pairings, _ := d.tourney.GeneratePairings()
	return pairings
}

func (d *Director) CompleteDraftAndEnterBattle() {
	log.Printf("[Round %d] Client checked in as READY. Simulating autonomous drafting brackets...", d.tourney.CurrentRound())

	var whiteCompetitor, blackCompetitor *game.PlayerProfile
	activePairings := d.getCurrentTourneyPairings()

	for _, pair := range activePairings {
		if pair.WhitePlayer.PlayerID == d.localClientID || pair.BlackPlayer.PlayerID == d.localClientID {
			whiteCompetitor = pair.WhitePlayer
			blackCompetitor = pair.BlackPlayer
			break
		}
	}

	globalBoard, _, err := game.MergePlayerPlacements(whiteCompetitor, blackCompetitor)
	if err != nil {
		log.Printf("Merge Phase Interception Error: %v", err)
	}

	log.Printf("Draft configuration unified!  Transitioning local client to game phase.")

	matchState := &game.MatchState{
		Board:       globalBoard,
		WhitePlayer: whiteCompetitor,
		BlackPlayer: blackCompetitor,
		ActiveColor: "white",
	}

	bScene, err := NewBattleScene("assets/engines/stockfish", matchState)
	d.currentScene = bScene
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
