package game

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MatchMergeTestSuite struct {
	suite.Suite
}

func TestRunMergeSuite(t *testing.T) {
	suite.Run(t, new(MatchMergeTestSuite))
}

func (s *MatchMergeTestSuite) TestMergePlayerPlacements_TableDriven() {
	tests := []struct {
		name            string
		setupWhite      func(p *PlayerProfile)
		setupBlack      func(p *PlayerProfile)
		expectedSuccess bool
		verifyBoard     func(b BoardState)
	}{
		{
			name: "Happy Path - Standard Valid Layout Merges Perfectly",
			setupWhite: func(p *PlayerProfile) {
				(*p.BoardAndBench.Board)[Location{File: 5, Rank: 1}] = Piece{Type: King, Color: White}
				(*p.BoardAndBench.Board)[Location{File: 4, Rank: 2}] = Piece{Type: Pawn, Color: White} // White d2
			},
			setupBlack: func(p *PlayerProfile) {
				(*p.BoardAndBench.Board)[Location{File: 5, Rank: 1}] = Piece{Type: King, Color: Black} // Local e1 -> Transposes to global d8
				(*p.BoardAndBench.Board)[Location{File: 4, Rank: 2}] = Piece{Type: Rook, Color: Black} // Local d2 -> Transposes to global e7
			},
			expectedSuccess: true,
			verifyBoard: func(b BoardState) {
				// Verify White piece positions are unaffected by flips
				pawn, pawnExists := b[Location{File: 4, Rank: 2}] // d2
				s.True(pawnExists)
				s.Equal(Pawn, pawn.Type)
				s.Equal(White, pawn.Color)

				// Verify Black piece positions mapped accurately into upper battlefield ranks via Transform()
				rook, rookExists := b[Location{File: 5, Rank: 7}] // e7
				s.True(rookExists)
				s.Equal(Rook, rook.Type)
				s.Equal(Black, rook.Color)
			},
		},
		{
			name: "Forgotten Kings - Automatic Deployment Fail-safe From Bench",
			setupWhite: func(p *PlayerProfile) {
				// Clear board map completely, leaving the starting King sitting isolated on the bench
				*p.BoardAndBench.Board = make(BoardState)
			},
			setupBlack: func(p *PlayerProfile) {
				*p.BoardAndBench.Board = make(BoardState)
			},
			expectedSuccess: true,
			verifyBoard: func(b BoardState) {
				// The merge loop must find the bench kings and auto-deploy them somewhere valid on the board grid
				whiteKingFound := false
				blackKing集中 := false
				for _, piece := range b {
					if piece.Type == King {
						if piece.Color == White {
							whiteKingFound = true
						} else if piece.Color == Black {
							blackKing集中 = true
						}
					}
				}
				s.True(whiteKingFound, "Failsafe referee failed to deploy forgotten White King from the bench array")
				s.True(blackKing集中, "Failsafe referee failed to deploy forgotten Black King from the bench array")
			},
		},
		{
			name: "Pawn Protection - Relocated Offending Pawns Never Land On Back Ranks",
			setupWhite: func(p *PlayerProfile) {
				(*p.BoardAndBench.Board)[Location{File: 5, Rank: 1}] = Piece{Type: King, Color: White}
				// Force an illegal position where an active check happens (White King e1 vs Black Rook e2)
				(*p.BoardAndBench.Board)[Location{File: 5, Rank: 2}] = Piece{Type: Pawn, Color: White} // White pawn on e2 blocking
			},
			setupBlack: func(p *PlayerProfile) {
				(*p.BoardAndBench.Board)[Location{File: 5, Rank: 1}] = Piece{Type: King, Color: Black}
				(*p.BoardAndBench.Board)[Location{File: 5, Rank: 7}] = Piece{Type: Rook, Color: Black} // Local e7 -> Global d2 checking White King on e1!
			},
			expectedSuccess: true,
			verifyBoard: func(b BoardState) {

				// Verify that no resulting pawns were accidentally scattered onto forbidden ranks 1 or 8
				for loc, piece := range b {
					if piece.Type == Pawn {
						s.NotEqual(1, loc.Rank, "Pawn scatter optimization dropped an asset onto forbidden back Rank 1")
						s.NotEqual(8, loc.Rank, "Pawn scatter optimization dropped an asset onto forbidden back Rank 8")
					}
				}
			},
		},
		{
			name: "Severe Deadlock Layout - Triggering Total 2-King Fallback Baseline Safely",
			setupWhite: func(p *PlayerProfile) {
				(*p.BoardAndBench.Board)[Location{File: 5, Rank: 1}] = Piece{Type: King, Color: White}

				for file := 1; file <= 8; file++ {
					(*p.BoardAndBench.Board)[Location{File: file, Rank: 2}] = Piece{Type: Pawn, Color: White}
				}
			},
			setupBlack: func(p *PlayerProfile) {
				// We deliberately do NOT deploy a King for Black, and we clear their bench completely.
				// This makes it impossible for ensureKingIsDeployed or spawnKingOnRandomOpenSquare
				// to ever locate a valid King asset, forcing the engine past the 50 pass limit.
				*p.BoardAndBench.Board = make(BoardState)
				p.BoardAndBench.Bench = []Piece{}
			},
			expectedSuccess: false, // This will successfully force the 50-pass failure and trigger the fallback
			verifyBoard: func(b BoardState) {
				s.Len(b, 2, "Nuclear fallback failed to purge broken grid components down to exactly 2 items")

				wKing, wExists := b[Location{File: 5, Rank: 1}]
				s.True(wExists)
				s.Equal(King, wKing.Type)
				s.Equal(White, wKing.Color)

				bKing, bExists := b[Location{File: 5, Rank: 8}]
				s.True(bExists)
				s.Equal(King, bKing.Type)
				s.Equal(Black, bKing.Color)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			whiteProfile := NewDefaultProfile("white_player", false)
			blackProfile := NewDefaultProfile("black_player", false)

			// Execute mock configurations
			tt.setupWhite(whiteProfile)
			tt.setupBlack(blackProfile)

			globalBoard, success, err := MergePlayerPlacements(whiteProfile, blackProfile)

			s.NoError(err)
			s.Equal(tt.expectedSuccess, success, "Merge status success flag parameter mismatch")
			s.NotNil(globalBoard)
			tt.verifyBoard(globalBoard)
		})
	}
}
