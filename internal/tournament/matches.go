package tournament

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/CoupDeGrace92/candidates/internal/game"
)

type ConditionType string

const (
	ConditionFixedRounds ConditionType = "FixedRounds"
	ConditionFirstToWins ConditionType = "FirstToWins"
	ConditionFirstToPts  ConditionType = "FirstToPoints"
)

type WinCondition struct {
	Type  ConditionType
	Limit float64
}

type MatchTournament struct {
	condition       WinCondition
	currentRoundNum int
	completed       bool

	winGold         int
	lossGold        int
	drawGold        int
	maxInterestGold int
	interestRate    int

	playerA *game.PlayerProfile
	playerB *game.PlayerProfile

	recA *ParticipantRecord
	recB *ParticipantRecord

	activePairings []Pairing
}

func NewMatchTournament(cond WinCondition, winG, lossG, drawG, interest, maxInterest int) *MatchTournament {
	return &MatchTournament{
		condition:       cond,
		currentRoundNum: 1,
		winGold:         winG,
		lossGold:        lossG,
		drawGold:        drawG,
		maxInterestGold: maxInterest,
		interestRate:    interest,
	}
}

func (m *MatchTournament) Initialize(participants []*game.PlayerProfile) {
	if len(participants) < 2 {
		return
	}

	m.playerA = participants[0]
	m.playerB = participants[1]

	m.recA = &ParticipantRecord{
		PlayerID:      m.playerA.PlayerID,
		ScoresByRound: make(map[int]float64),
	}
	m.recB = &ParticipantRecord{
		PlayerID:      m.playerB.PlayerID,
		ScoresByRound: make(map[int]float64),
	}
}

func (m *MatchTournament) CurrentRound() int {
	return m.currentRoundNum
}

func (m *MatchTournament) IsCompleted() bool {
	return m.completed
}

func (m *MatchTournament) GeneratePairings() ([]Pairing, error) {
	if m.completed {
		return nil, fmt.Errorf("tournament is already completed!")
	}

	m.activePairings = []Pairing{}

	var useAAsWhite bool
	if m.currentRoundNum == 1 {
		nBig, _ := rand.Int(rand.Reader, big.NewInt(2))
		useAAsWhite = nBig.Int64() == 0
	} else {
		useAAsWhite = m.currentRoundNum%2 != 0
	}

	var white, black *game.PlayerProfile
	if useAAsWhite {
		white = m.playerA
		black = m.playerB
	} else {
		white = m.playerB
		black = m.playerA
	}

	pID := fmt.Sprintf("p_r%d_%s_vs_%s", m.currentRoundNum, white.PlayerID, black.PlayerID)
	p := Pairing{
		ID:          pID,
		WhitePlayer: white,
		BlackPlayer: black,
		RoundNum:    m.currentRoundNum,
		Game:        GameContainer{IsPlayed: false, PGN: ""},
	}

	m.activePairings = append(m.activePairings, p)
	return m.activePairings, nil
}

func (m *MatchTournament) ResolveMatchOutcome(outcome MatchOutcome) {
	if len(m.activePairings) == 0 || m.activePairings[0].ID != outcome.PairingID {
		return //Stale or mismatched execution envelope
	}

	activePair := m.activePairings[0]
	activePair.Game.IsPlayed = true
	activePair.Game.PGN = outcome.PGNData

	m.recA.History = append(m.recA.History, activePair)
	m.recB.History = append(m.recB.History, activePair)

	m.applyRoundInterest(m.playerA)
	m.applyRoundInterest(m.playerB)

	if outcome.IsDraw {
		m.recA.Draws++
		m.recA.Points += 0.5
		m.recA.ScoresByRound[m.currentRoundNum] = 0.5

		m.recB.Draws++
		m.recB.Points += 0.5
		m.recB.ScoresByRound[m.currentRoundNum] = 0.5

		m.playerA.Gold += m.drawGold
		m.playerB.Gold += m.drawGold
	} else {
		if outcome.WinnerID == m.playerA.PlayerID {
			m.recA.Wins++
			m.recA.Points += 1.0
			m.recA.ScoresByRound[m.currentRoundNum] = 1.0

			m.recB.Losses++
			m.recB.ScoresByRound[m.currentRoundNum] = 0.0

			m.playerA.Gold += m.winGold
			m.playerB.Gold += m.lossGold
		} else {
			m.recB.Wins++
			m.recB.Points += 1.0
			m.recB.ScoresByRound[m.currentRoundNum] = 1.0

			m.recA.Losses++
			m.recA.ScoresByRound[m.currentRoundNum] = 0.0

			m.playerB.Gold += m.winGold
			m.playerA.Gold += m.lossGold
		}
	}

	switch m.condition.Type {
	case ConditionFixedRounds:
		if float64(m.currentRoundNum) >= m.condition.Limit {
			m.completed = true
		}
	case ConditionFirstToWins:
		targetWins := int(m.condition.Limit)
		if m.recA.Wins >= targetWins || m.recB.Wins >= targetWins {
			m.completed = true
		}
	case ConditionFirstToPts:
		targetPoints := float64(m.condition.Limit)
		if m.recA.Points >= targetPoints || m.recB.Points >= targetPoints {
			m.completed = true
		}
	}
	if !m.completed {
		m.currentRoundNum++
	}
}

func (m *MatchTournament) applyRoundInterest(p *game.PlayerProfile) {
	if p == nil {
		return
	}
	interestEarned := p.Gold * m.interestRate
	if interestEarned > m.maxInterestGold {
		interestEarned = m.maxInterestGold
	}
	p.Gold += interestEarned
}

func (m *MatchTournament) GetCrossTable() []ParticipantRecord {
	list := []ParticipantRecord{*m.recA, *m.recB}
	if list[1].Points > list[0].Points {
		list[0], list[1] = list[1], list[0]
	}
	return list
}
