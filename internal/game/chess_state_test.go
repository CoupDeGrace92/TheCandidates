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
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 5, Rank: 1}: {Type: King, Color: White}, // E1
							Location{File: 1, Rank: 1}: {Type: Rook, Color: White}, // A1
							Location{File: 8, Rank: 1}: {Type: Rook, Color: White}, // H1
						},
					},
				},
				BlackPlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 5, Rank: 8}: {Type: King, Color: Black}, // E8
							Location{File: 1, Rank: 8}: {Type: Rook, Color: Black}, // A8
							Location{File: 8, Rank: 8}: {Type: Rook, Color: Black}, // H8
						},
					},
				},
				ActiveColor:     "white",
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
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 5, Rank: 1}: {Type: King, Color: White}, // E1
							Location{File: 8, Rank: 1}: {Type: Rook, Color: White}, // H1
						},
					},
				},
				BlackPlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 5, Rank: 8}: {Type: King, Color: Black}, // E8
							Location{File: 1, Rank: 8}: {Type: Rook, Color: Black}, // A8
						},
					},
				},
				ActiveColor:     "black",
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
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 5, Rank: 1}: {Type: King, Color: White},
						},
					},
				},
				BlackPlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 5, Rank: 8}: {Type: King, Color: Black},
						},
					},
				},
				ActiveColor:     "white",
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
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 2, Rank: 2}: {Type: Pawn, Color: White}, // B2
							Location{File: 4, Rank: 3}: {Type: King, Color: White}, // D3
						},
					},
				},
				BlackPlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 7, Rank: 7}: {Type: Pawn, Color: Black}, // G7
							Location{File: 5, Rank: 5}: {Type: King, Color: Black}, // E5
						},
					},
				},
				ActiveColor:     "black",
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
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 5, Rank: 1}: {Type: King, Color: White},   // E1
							Location{File: 3, Rank: 5}: {Type: Bishop, Color: White}, // C5
							Location{File: 4, Rank: 5}: {Type: Pawn, Color: White},   // D5
							Location{File: 1, Rank: 1}: {Type: Rook, Color: White},   // A1
						},
					},
				},
				BlackPlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 5, Rank: 8}: {Type: King, Color: Black},   // E8
							Location{File: 5, Rank: 5}: {Type: Pawn, Color: Black},   // E5
							Location{File: 6, Rank: 6}: {Type: Knight, Color: Black}, // F6
						},
					},
				},
				ActiveColor:     "white",
				CastlingRights:  "Q",
				EnPassantTarget: "e6",
				HalfMoveClock:   0,
				FullMoveNumber:  12,
			},
			expectedFEN: "4k3/8/5n2/2BPp3/8/8/8/R3K3 w Q e6 0 12",
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
				(*m.WhitePlayer.BoardAndBench.Board)[Location{File: 5, Rank: 1}] = Piece{Type: King, Color: White}
				(*m.WhitePlayer.BoardAndBench.Board)[Location{File: 1, Rank: 1}] = Piece{Type: Rook, Color: White}
				(*m.WhitePlayer.BoardAndBench.Board)[Location{File: 8, Rank: 1}] = Piece{Type: Rook, Color: White}

				(*m.BlackPlayer.BoardAndBench.Board)[Location{File: 5, Rank: 8}] = Piece{Type: King, Color: Black}
				(*m.BlackPlayer.BoardAndBench.Board)[Location{File: 1, Rank: 8}] = Piece{Type: Rook, Color: Black}
				(*m.BlackPlayer.BoardAndBench.Board)[Location{File: 8, Rank: 8}] = Piece{Type: Rook, Color: Black}
			},
			expectedRights: "KQkq",
		},
		{
			name: "Asymmetric Rights",
			setupBoard: func(m *MatchState) {
				(*m.WhitePlayer.BoardAndBench.Board)[Location{File: 4, Rank: 1}] = Piece{Type: King, Color: White}
				(*m.WhitePlayer.BoardAndBench.Board)[Location{File: 8, Rank: 1}] = Piece{Type: Rook, Color: White}

				(*m.BlackPlayer.BoardAndBench.Board)[Location{File: 5, Rank: 8}] = Piece{Type: King, Color: Black}
				(*m.BlackPlayer.BoardAndBench.Board)[Location{File: 8, Rank: 8}] = Piece{Type: Rook, Color: Black}
			},
			expectedRights: "k",
		},
		{
			name: "No Rights Available",
			setupBoard: func(m *MatchState) {
				(*m.WhitePlayer.BoardAndBench.Board)[Location{File: 5, Rank: 5}] = Piece{Type: King, Color: White}
				(*m.BlackPlayer.BoardAndBench.Board)[Location{File: 5, Rank: 6}] = Piece{Type: King, Color: Black}
			},
			expectedRights: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wbs := make(BoardState)
			bbs := make(BoardState)
			ms := &MatchState{
				WhitePlayer: &PlayerProfile{BoardAndBench: &PlayerPieces{Board: &wbs}},
				BlackPlayer: &PlayerProfile{BoardAndBench: &PlayerPieces{Board: &bbs}},
			}
			tt.setupBoard(ms)

			ms.InitializeCastlingRights()

			assert.Equal(t, tt.expectedRights, ms.CastlingRights)
		})
	}
}

func TestMatchState_ApplyMove(t *testing.T) {
	tests := []struct {
		name          string
		initialState  *MatchState
		move          string
		verifyResults func(t *testing.T, ms *MatchState)
	}{
		{
			name: "Standard Pawn Push Updates Clocks and En Passant Target",
			initialState: &MatchState{
				ActiveColor:    "white",
				CastlingRights: "-",
				WhitePlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{Location{File: 5, Rank: 2}: {Type: Pawn, Color: White}}, // e2
					},
				},
				BlackPlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{},
					},
				},
			},
			move: "e2e4",
			verifyResults: func(t *testing.T, ms *MatchState) {
				assert.Equal(t, Black, ms.ActiveColor)
				assert.Equal(t, "e3", ms.EnPassantTarget)
				assert.Equal(t, 0, ms.HalfMoveClock)
				_, exists := (*ms.WhitePlayer.BoardAndBench.Board)[Location{File: 5, Rank: 4}] // e4
				assert.True(t, exists)
			},
		},
		{
			name: "King Move Degrades Castling Rights Perfectly",
			initialState: &MatchState{
				ActiveColor:    "white",
				CastlingRights: "KQkq",
				WhitePlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{Location{File: 5, Rank: 1}: {Type: King, Color: White}}, // e1
					},
				},
				BlackPlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{},
					},
				},
			},
			move: "e1d1",
			verifyResults: func(t *testing.T, ms *MatchState) {
				assert.Equal(t, "kq", ms.CastlingRights)
			},
		},
		{
			name: "White Kingside Castling Correctly Repositions King and Rook",
			initialState: &MatchState{
				ActiveColor:    "white",
				CastlingRights: "KQkq",
				WhitePlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{
							Location{File: 5, Rank: 1}: {Type: King, Color: White}, // e1
							Location{File: 8, Rank: 1}: {Type: Rook, Color: White}, // h1
						},
					},
				},
				BlackPlayer: &PlayerProfile{
					BoardAndBench: &PlayerPieces{
						Board: &BoardState{},
					},
				},
			},
			move: "e1g1",
			verifyResults: func(t *testing.T, ms *MatchState) {
				assert.Equal(t, "kq", ms.CastlingRights)
				_, kingExists := (*ms.WhitePlayer.BoardAndBench.Board)[Location{File: 7, Rank: 1}]
				_, rookExists := (*ms.WhitePlayer.BoardAndBench.Board)[Location{File: 6, Rank: 1}]
				assert.True(t, kingExists && rookExists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.initialState.ApplyMove(tt.move)
			assert.NoError(t, err)
			tt.verifyResults(t, tt.initialState)
		})
	}
}

func TestPlayerPieces_BenchToBoard(t *testing.T) {
	targetLoc := Location{File: 4, Rank: 3}      // d3
	testPiece := Piece{Type: Pawn, Color: White} // Removed ID string

	p := NewPlayerPieces()
	p.Bench = append(p.Bench, testPiece)

	lockedLoc := Location{File: 4, Rank: 5}
	successLocked := p.BenchToBoard(0, lockedLoc)
	assert.False(t, successLocked)
	assert.Len(t, p.Bench, 1)

	p.Squares[targetLoc] = struct{}{}
	successValid := p.BenchToBoard(0, targetLoc)

	assert.True(t, successValid)
	assert.Len(t, p.Bench, 0)

	boardPiece, onBoard := (*p.Board)[targetLoc]
	assert.True(t, onBoard)
	assert.Equal(t, Pawn, boardPiece.Type)

	p.Bench = append(p.Bench, Piece{Type: Knight, Color: White})
	successCollision := p.BenchToBoard(0, targetLoc)
	assert.False(t, successCollision)
}

func TestPlayerPieces_BoardToBench(t *testing.T) {
	targetLoc := Location{File: 5, Rank: 2} // e2
	p := NewPlayerPieces()
	testPiece := Piece{Type: Rook, Color: White}
	(*p.Board)[targetLoc] = testPiece

	emptyLoc := Location{File: 1, Rank: 1}
	successEmpty := p.BoardToBench(emptyLoc)

	assert.False(t, successEmpty)

	// Recall the valid Rook from e2 back to the bench inventory
	initialBenchLen := len(p.Bench)
	successRecall := p.BoardToBench(targetLoc)

	assert.True(t, successRecall)
	assert.Len(t, p.Bench, initialBenchLen+1)
	assert.Equal(t, Rook, p.Bench[initialBenchLen].Type)

	_, onBoard := (*p.Board)[targetLoc]
	assert.False(t, onBoard)
}

func TestPlayerPieces_BoardToBoard(t *testing.T) {
	startLoc := Location{File: 5, Rank: 2}  // e2
	validLoc := Location{File: 5, Rank: 3}  // e3
	lockedLoc := Location{File: 5, Rank: 5} // e5

	p := NewPlayerPieces()
	testPiece := Piece{Type: Pawn, Color: White}
	(*p.Board)[startLoc] = testPiece

	successLocked := p.BoardToBoard(startLoc, lockedLoc)
	assert.False(t, successLocked)

	emptyLoc := Location{File: 1, Rank: 1}
	successEmpty := p.BoardToBoard(emptyLoc, validLoc)
	assert.False(t, successEmpty)

	p.Squares[validLoc] = struct{}{}
	successValid := p.BoardToBoard(startLoc, validLoc)

	assert.True(t, successValid)

	_, oldOccupied := (*p.Board)[startLoc]
	assert.False(t, oldOccupied)

	boardPiece, newOccupied := (*p.Board)[validLoc]
	assert.True(t, newOccupied)
	assert.Equal(t, Pawn, boardPiece.Type)
}
