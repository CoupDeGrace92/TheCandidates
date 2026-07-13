package tournament

import "github.com/CoupDeGrace92/candidates/internal/game"

// THIS IS THE IMPORTANT INTERFACE THAT DETERMINES HOW A TOURNAMENT FUNCTIONS
type TournamentManager interface {
	Initialize(participants []*game.PlayerProfile)
	CurrentRound() int
	IsCompleted() bool //True if the last round has finished and a winner is chosen
	GeneratePairings() ([]Pairing, error)
	ResolveGameOutcome(outcome MatchOutcome)
	GetCrossTable() []ParticipantRecord
}

type MatchOutcome struct {
	PairingID string
	PlayerAID string
	PlayerBID string
	WinnerID  string
	IsDraw    bool
	PGNData   string
}

type ParticipantRecord struct {
	PlayerID      string
	Wins          int
	Losses        int
	Draws         int
	Points        float64
	History       []Pairing
	ScoresByRound map[int]float64
}

type Pairing struct {
	ID          string //unique string for mapping lookups
	WhitePlayer *game.PlayerProfile
	BlackPlayer *game.PlayerProfile
	RoundNum    int
	Game        GameContainer //Currently not implemented but will contain the PGN of the game played
}

type GameContainer struct {
	IsPlayed bool
	PGN      string //Currently not implemented
}
