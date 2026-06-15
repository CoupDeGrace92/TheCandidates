package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcatenateBoardState(t *testing.T) {
	posWhite := Location{Rank: 1, File: 5}
	posBlack := Location{Rank: 8, File: 5}

	whiteBoard := BoardState{
		posWhite: {Type: King, Color: White},
	}
	blackBoard := BoardState{
		posBlack: {Type: King, Color: Black},
	}

	merged := ConcatenateBoardState(&whiteBoard, &blackBoard)

	require.NotNil(t, merged)
	assert.Len(t, *merged, 2)
	assert.Equal(t, King, (*merged)[posWhite].Type)
	assert.Equal(t, White, (*merged)[posWhite].Color)
	assert.Equal(t, King, (*merged)[posBlack].Type)
	assert.Equal(t, Black, (*merged)[posBlack].Color)
}

func TestMatchState_ToFEN_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		state       *MatchState
		expectedFEN string
	}{
		{
			name: "Full Castling Rights",
			state: &MatchState{
				WhitePlayer: &PlayerProfile{
					Board: &BoardState{
						Location{File: 5, Rank: 1}: {Type: King, Color: White}, // E1
						Location{File: 1, Rank: 1}: {Type: Rook, Color: White}, // A1
						Location{File: 8, Rank: 1}: {Type: Rook, Color: White}, // H1
					},
				},
				BlackPlayer: &PlayerProfile{
					Board: &BoardState{
						Location{File: 5, Rank: 8}: {Type: King, Color: Black}, // E8
						Location{File: 1, Rank: 8}: {Type: Rook, Color: Black}, // A8
						Location{File: 8, Rank: 8}: {Type: Rook, Color: Black}, // H8
					},
				},
				ActiveColor:     "w",
				CastlingRights:  "KQkq",
				EnPassantTarget: "-",
				HalfMoveClock:   0,
				FullMoveNumber:  1,
			},
			expectedFEN: "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
		},
		{
			name: "Partial Castling Rights",
			state: &MatchState{
				WhitePlayer: &PlayerProfile{
					Board: &BoardState{
						Location{File: 5, Rank: 1}: {Type: King, Color: White}, // E1
						Location{File: 8, Rank: 1}: {Type: Rook, Color: White}, // H1
					},
				},
				BlackPlayer: &PlayerProfile{
					Board: &BoardState{
						Location{File: 5, Rank: 8}: {Type: King, Color: Black}, // E8
						Location{File: 1, Rank: 8}: {Type: Rook, Color: Black}, // A8
					},
				},
				ActiveColor:     "b",
				CastlingRights:  "Kq",
				EnPassantTarget: "-",
				HalfMoveClock:   2,
				FullMoveNumber:  5,
			},
			expectedFEN: "r3k3/8/8/8/8/8/8/4K2R b Kq - 2 5",
		},
		{
			name: "Half Move and Full Move Greater Than Zero",
			state: &MatchState{
				WhitePlayer: &PlayerProfile{
					Board: &BoardState{
						Location{File: 5, Rank: 1}: {Type: King, Color: White},
					},
				},
				BlackPlayer: &PlayerProfile{
					Board: &BoardState{
						Location{File: 5, Rank: 8}: {Type: King, Color: Black},
					},
				},
				ActiveColor:     "w",
				CastlingRights:  "-",
				EnPassantTarget: "-",
				HalfMoveClock:   14,
				FullMoveNumber:  22,
			},
			expectedFEN: "4k3/8/8/8/8/8/8/4K3 w - - 14 22",
		},
		{
			name: "No EP or Castling Rights",
			state: &MatchState{
				WhitePlayer: &PlayerProfile{
					Board: &BoardState{
						Location{File: 2, Rank: 2}: {Type: Pawn, Color: White}, // B2
						Location{File: 4, Rank: 3}: {Type: King, Color: White}, // D3
					},
				},
				BlackPlayer: &PlayerProfile{
					Board: &BoardState{
						Location{File: 7, Rank: 7}: {Type: Pawn, Color: Black}, // G7
						Location{File: 5, Rank: 5}: {Type: King, Color: Black}, // E5 (Fixed from Rank 6 to match expected FEN)
					},
				},
				ActiveColor:     "b",
				CastlingRights:  "-",
				EnPassantTarget: "-",
				HalfMoveClock:   0,
				FullMoveNumber:  8,
			},
			expectedFEN: "8/6p1/8/4k3/8/3K4/1P6/8 b - - 0 8",
		},
		{
			name: "Semi Complex Position (With Active EP Target)",
			state: &MatchState{
				WhitePlayer: &PlayerProfile{
					Board: &BoardState{
						Location{File: 5, Rank: 1}: {Type: King, Color: White},   // E1
						Location{File: 3, Rank: 5}: {Type: Bishop, Color: White}, // C5
						Location{File: 4, Rank: 5}: {Type: Pawn, Color: White},   // D5
						Location{File: 1, Rank: 1}: {Type: Rook, Color: White},   // A1
					},
				},
				BlackPlayer: &PlayerProfile{
					Board: &BoardState{
						Location{File: 5, Rank: 8}: {Type: King, Color: Black},   // E8
						Location{File: 5, Rank: 5}: {Type: Pawn, Color: Black},   // E5
						Location{File: 6, Rank: 6}: {Type: Knight, Color: Black}, // F6
					},
				},
				ActiveColor:     "w",
				CastlingRights:  "Q",
				EnPassantTarget: "e6",
				HalfMoveClock:   0,
				FullMoveNumber:  12,
			},
			expectedFEN: "4k3/8/5n2/2BPp3/8/8/8/R3K3 w Q e6 0 12", // Fixed expected 'b' to uppercase 'B'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualFEN := tt.state.ToFEN()
			assert.Equal(t, tt.expectedFEN, actualFEN)
		})
	}
}

func TestMatchState_InitializeCastlingRights(t *testing.T) {
	tests := []struct {
		name           string
		setupBoard     func(ms *MatchState)
		expectedRights string
	}{
		{
			name: "Full Rights Intact",
			setupBoard: func(m *MatchState) {
				(*m.WhitePlayer.Board)[Location{File: 5, Rank: 1}] = Piece{Type: King, Color: White}
				(*m.WhitePlayer.Board)[Location{File: 1, Rank: 1}] = Piece{Type: Rook, Color: White}
				(*m.WhitePlayer.Board)[Location{File: 8, Rank: 1}] = Piece{Type: Rook, Color: White}

				(*m.BlackPlayer.Board)[Location{File: 5, Rank: 8}] = Piece{Type: King, Color: Black}
				(*m.BlackPlayer.Board)[Location{File: 1, Rank: 8}] = Piece{Type: Rook, Color: Black}
				(*m.BlackPlayer.Board)[Location{File: 8, Rank: 8}] = Piece{Type: Rook, Color: Black}
			},
			expectedRights: "KQkq",
		},
		{
			name: "Asymmetric Rights",
			setupBoard: func(m *MatchState) {
				(*m.WhitePlayer.Board)[Location{File: 4, Rank: 1}] = Piece{Type: King, Color: White}
				(*m.WhitePlayer.Board)[Location{File: 8, Rank: 1}] = Piece{Type: Rook, Color: White}

				(*m.BlackPlayer.Board)[Location{File: 5, Rank: 8}] = Piece{Type: King, Color: Black}
				(*m.BlackPlayer.Board)[Location{File: 8, Rank: 8}] = Piece{Type: Rook, Color: Black}
			},
			expectedRights: "k",
		},
		{
			name: "No Rights Available",
			setupBoard: func(m *MatchState) {
				(*m.WhitePlayer.Board)[Location{File: 5, Rank: 5}] = Piece{Type: King, Color: White}
				(*m.BlackPlayer.Board)[Location{File: 5, Rank: 6}] = Piece{Type: King, Color: Black}
			},
			expectedRights: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wbs := make(BoardState)
			bbs := make(BoardState)
			ms := &MatchState{
				WhitePlayer: &PlayerProfile{Board: &wbs},
				BlackPlayer: &PlayerProfile{Board: &bbs},
			}
			tt.setupBoard(ms)

			ms.InitializeCastlingRights()

			assert.Equal(t, tt.expectedRights, ms.CastlingRights)
		})
	}
}
