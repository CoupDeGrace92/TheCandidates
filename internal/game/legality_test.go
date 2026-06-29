package game

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LegalityChecksTestSuite struct {
	suite.Suite
}

func TestRunLegalitySuite(t *testing.T) {
	suite.Run(t, new(LegalityChecksTestSuite))
}

func (s *LegalityChecksTestSuite) TestLegalPosition_TableDriven() {
	tests := []struct {
		name              string
		activeColor       PieceColor
		setupBoard        func() *BoardState
		expectedLegal     bool
		expectedWK        bool
		expectedBK        bool
		expectedOffending int // Assert how many coordinates should be flagged
	}{
		{
			name:        "Happy Path - Standard Valid Position",
			activeColor: White,
			setupBoard: func() *BoardState {
				return &BoardState{
					Location{File: 5, Rank: 1}: {Type: King, Color: White},
					Location{File: 5, Rank: 8}: {Type: King, Color: Black},
					Location{File: 5, Rank: 2}: {Type: Pawn, Color: White},
				}
			},
			expectedLegal:     true,
			expectedWK:        true,
			expectedBK:        true,
			expectedOffending: 0,
		},
		{
			name:        "Catastrophic Error - Missing White King Completely",
			activeColor: White,
			setupBoard: func() *BoardState {
				return &BoardState{
					Location{File: 5, Rank: 8}: {Type: King, Color: Black},
				}
			},
			expectedLegal:     false,
			expectedWK:        false, // Flagged missing
			expectedBK:        true,
			expectedOffending: 0,
		},
		{
			name:        "Catastrophic Error - Missing Black King Completely",
			activeColor: White,
			setupBoard: func() *BoardState {
				return &BoardState{
					Location{File: 5, Rank: 1}: {Type: King, Color: White},
				}
			},
			expectedLegal:     false,
			expectedWK:        true,
			expectedBK:        false, // Flagged missing
			expectedOffending: 0,
		},
		{
			name:        "Illegal Layout - Duplicate White King Flagged",
			activeColor: White,
			setupBoard: func() *BoardState {
				return &BoardState{
					Location{File: 5, Rank: 1}: {Type: King, Color: White},
					Location{File: 6, Rank: 1}: {Type: King, Color: White}, // Duplicate
					Location{File: 5, Rank: 8}: {Type: King, Color: Black},
				}
			},
			expectedLegal:     false,
			expectedWK:        false, // The function flips this to false for any violation
			expectedBK:        true,
			expectedOffending: 1, // Flips legal to false and saves coordinate (6,1)
		},
		{
			name:        "Illegal Layout - White Pawn Trapped on Back Row Rank 1",
			activeColor: White,
			setupBoard: func() *BoardState {
				return &BoardState{
					Location{File: 5, Rank: 1}: {Type: King, Color: White},
					Location{File: 5, Rank: 8}: {Type: King, Color: Black},
					Location{File: 3, Rank: 1}: {Type: Pawn, Color: White}, // Illegal coordinate
				}
			},
			expectedLegal:     false,
			expectedWK:        true,
			expectedBK:        true,
			expectedOffending: 1, // Flags coordinate (3,1)
		},
		{
			name:        "Illegal Sequence - Inactive Black King Left In Check by White Rook",
			activeColor: White,
			setupBoard: func() *BoardState {
				return &BoardState{
					Location{File: 5, Rank: 1}: {Type: King, Color: White},
					Location{File: 5, Rank: 8}: {Type: King, Color: Black}, // Inactive King at e8
					Location{File: 1, Rank: 8}: {Type: Rook, Color: White}, // a8 lines up e8
				}
			},
			expectedLegal:     false,
			expectedWK:        true,
			expectedBK:        true,
			expectedOffending: 1, // Flags the inactive King location (5,8) as the offense
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			board := tt.setupBoard()

			legal, wk, bk, offending := LegalPosition(board, tt.activeColor)

			s.Equal(tt.expectedLegal, legal, "Legal parameter mismatch")
			s.Equal(tt.expectedWK, wk, "White King tracking parameter mismatch")
			s.Equal(tt.expectedBK, bk, "Black King tracking parameter mismatch")
			s.Len(offending, tt.expectedOffending, "Offending positions slice headcount mismatch")
		})
	}
}
