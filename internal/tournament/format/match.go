package formats

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/CoupDeGrace92/candidates/internal/game"
	"github.com/CoupDeGrace92/candidates/internal/tournament"
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
	interestRate    float64
	maxInterestGold int

	playerA *game.PlayerProfile
	playerB *game.PlayerProfile

	recA *tournament.ParticipantRecord
	recB *tournament.ParticipantRecord

	CurrentPairings []tournament.Pairing
}

func NewMatchTournament(cond WinCondition, winG, lossG, drawG, maxInterest int, interest float64) *MatchTournament {
	return &MatchTournament{
		condition:       cond,
		currentRoundNum: 1,
		winGold:         winG,
		lossGold:        lossG,
		drawGold:        drawG,
		interestRate:    interest,
		maxInterestGold: maxInterest,

		CurrentPairings: []tournament.Pairing{},
	}
}

func (m *MatchTournament) Initialize(participants []*game.PlayerProfile) {
	if len(participants) != 2 {
		return
	}

	m.playerA = participants[0]
	m.playerB = participants[1]

	m.recA = &tournament.ParticipantRecord{PlayerID: m.playerA.PlayerID, ScoresByRound: make(map[int]float64)}
	m.recB = &tournament.ParticipantRecord{PlayerID: m.playerB.PlayerID, ScoresByRound: make(map[int]float64)}
}

func (m *MatchTournament) CurrentRound() int {
	return m.currentRoundNum
}

func (m *MatchTournament) IsCompleted() bool {
	return m.completed
}

func (m *MatchTournament) GeneratePairings() ([]tournament.Pairing, error) {
	if m.completed {
		return nil, fmt.Errorf("Match is already completed")
	}

	if len(m.CurrentPairings) > 0 && m.CurrentPairings[0].RoundNum == m.currentRoundNum {
		return m.CurrentPairings, nil
	}

	m.CurrentPairings = []tournament.Pairing{}

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

	p := tournament.Pairing{
		ID:          pID,
		WhitePlayer: white,
		BlackPlayer: black,
		RoundNum:    m.currentRoundNum,
		Game:        tournament.GameContainer{IsPlayed: false, PGN: ""},
	}

	m.CurrentPairings = append(m.CurrentPairings, p)
	return m.CurrentPairings, nil
}

func (m *MatchTournament) ResolveMatchOutcome(outcome tournament.MatchOutcome) {
	if len(m.CurrentPairings) == 0 || m.CurrentPairings[0].ID != outcome.PairingID {
		return
	}

	activePair := m.CurrentPairings[0]
	activePair.Game.IsPlayed = true
	activePair.Game.PGN = outcome.PGNData

	m.recA.History = append(m.recA.History, activePair)
	m.recB.History = append(m.recB.History, activePair)

	if outcome.IsDraw {
		m.recA.Draws++
		m.recA.Points += 0.5
		m.recA.ScoresByRound[m.currentRoundNum] = .5

		m.recB.Draws++
		m.recB.Points += 0.5
		m.recB.ScoresByRound[m.currentRoundNum] = .5

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
			m.recA.Losses++
			m.recA.ScoresByRound[m.currentRoundNum] = 0.0

			m.recB.Wins++
			m.recB.Points += 1.0
			m.recB.ScoresByRound[m.currentRoundNum] = 1.0

			m.playerA.Gold += m.lossGold
			m.playerB.Gold += m.winGold
		}
	}

	m.applyRoundInterest(m.playerA)
	m.applyRoundInterest(m.playerB)

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
		if lim := m.condition.Limit; m.recA.Points >= lim || m.recB.Points >= lim {
			m.completed = true
		}
	}

	if !m.completed {
		m.currentRoundNum++
		m.CurrentPairings = []tournament.Pairing{}
	}
}

func (m *MatchTournament) applyRoundInterest(p *game.PlayerProfile) {
	if p == nil {
		return
	}
	interestEarned := int(math.Floor(m.interestRate * float64(p.Gold)))
	if interestEarned > m.maxInterestGold {
		interestEarned = m.maxInterestGold
	}

	p.Gold += interestEarned
}

func (m *MatchTournament) GetCrossTable() []tournament.ParticipantRecord {
	list := []tournament.ParticipantRecord{*m.recA, *m.recB}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Points == list[j].Points {
			return list[i].Wins > list[j].Wins
		}
		return list[i].Points > list[j].Points
	})
	return list
}
