package game

import (
	"fmt"
	"strconv"
	"strings"
)

type PieceType string

const (
	Pawn   PieceType = "pawn"
	Knight PieceType = "knight"
	Bishop PieceType = "bishop"
	Rook   PieceType = "rook"
	Queen  PieceType = "queen"
	King   PieceType = "king"
)

type PieceColor string

const (
	White PieceColor = "white"
	Black PieceColor = "black"
)

type Piece struct {
	Type  PieceType  `json:"type"`
	Color PieceColor `json:"color"`
}

type Location struct {
	Rank int `json:"rank"`
	File int `json:"file"`
}

func (p *Location) IsValid() bool {
	return p.Rank > 0 && p.Rank <= 8 && p.File > 0 && p.File <= 8
}

type BoardState map[Location]Piece

type PlayerProfile struct {
	//This profile might move - this also might indicate which assets to pull for the player
	PlayerID string      `json:"player_id"`
	Health   int         `json:"health"`
	Gold     int         `json:"gold"`
	Board    *BoardState `json:"board"`
	Bench    []Piece     `json:"bench"`
	//Might also want - last white FEN and last black FEN
}

type MatchState struct {
	WhitePlayer     *PlayerProfile `json:"white_player"`
	BlackPlayer     *PlayerProfile `json:"black_player"`
	ActiveColor     PieceColor     `json:"active_color"`
	HalfMoveClock   int            `json:"halfmove_clock"`
	FullMoveNumber  int            `json:"fullmove_number"`
	CastlingRights  string         `json:"castling_rights"`
	EnPassantTarget string         `json:"en_passant_target"`
}

func ConcatenateBoardState(whiteBoard, blackBoard *BoardState) *BoardState {
	merged := make(BoardState)
	if whiteBoard != nil {
		for loc, piece := range *whiteBoard {
			merged[loc] = piece
		}
	}
	if blackBoard != nil {
		for loc, piece := range *blackBoard {
			merged[loc] = piece
		}
	}
	return &merged
}

func getPieceFENChar(p Piece) string {
	var char string
	switch p.Type {
	case Pawn:
		char = "p"
	case Knight:
		char = "n"
	case Bishop:
		char = "b"
	case Rook:
		char = "r"
	case Queen:
		char = "q"
	case King:
		char = "k"
	}
	if p.Color == "white" {
		return strings.ToUpper(char)
	}
	return char
}

func (m *MatchState) ToFEN() string {
	if m == nil {
		return ""
	}

	combinedBoard := ConcatenateBoardState(m.WhitePlayer.Board, m.BlackPlayer.Board)
	var rows []string
	//FEN reads top rank (Rank = 8) down to bottom rank (Rank = 0)
	for rank := 8; rank >= 1; rank-- {
		var rankStringBuilder strings.Builder
		emptyCount := 0

		for file := 1; file <= 8; file++ {
			loc := Location{Rank: rank, File: file}
			if piece, occupied := (*combinedBoard)[loc]; occupied {
				if emptyCount > 0 {
					rankStringBuilder.WriteString(strconv.Itoa(emptyCount))
					emptyCount = 0
				}
				rankStringBuilder.WriteString(getPieceFENChar(piece))
			} else {
				emptyCount++
			}
		}

		if emptyCount > 0 {
			rankStringBuilder.WriteString(strconv.Itoa(emptyCount))
		}
		rows = append(rows, rankStringBuilder.String())
	}

	fenBoard := strings.Join(rows, "/")

	castling := m.CastlingRights
	ep := m.EnPassantTarget
	if ep == "" {
		ep = "-"
	}
	activeColorChar := "w"
	if m.ActiveColor == Black {
		activeColorChar = "b"
	}

	return fmt.Sprintf("%s %s %s %s %d %d",
		fenBoard,
		activeColorChar,
		castling,
		ep,
		m.HalfMoveClock,
		m.FullMoveNumber,
	)
}

func (m *MatchState) InitializeCastlingRights() {
	rights := ""

	if wk, ok := (*m.WhitePlayer.Board)[Location{Rank: 1, File: 5}]; ok && wk.Type == King {
		if r, ok := (*m.WhitePlayer.Board)[Location{Rank: 1, File: 8}]; ok && r.Type == Rook {
			rights += "K"
		}
		if r, ok := (*m.WhitePlayer.Board)[Location{Rank: 1, File: 1}]; ok && r.Type == Rook {
			rights += "Q"
		}
	}

	if bk, ok := (*m.BlackPlayer.Board)[Location{Rank: 8, File: 5}]; ok && bk.Type == King {
		if r, ok := (*m.BlackPlayer.Board)[Location{Rank: 8, File: 8}]; ok && r.Type == Rook {
			rights += "k"
		}
		if r, ok := (*m.BlackPlayer.Board)[Location{Rank: 8, File: 1}]; ok && r.Type == Rook {
			rights += "q"
		}
	}

	if rights == "" {
		m.CastlingRights = "-"
	} else {
		m.CastlingRights = rights
	}
}

func (m *MatchState) ApplyMove(moveStr string) error {
	if len(moveStr) < 4 || len(moveStr) > 5 {
		return fmt.Errorf("invalid UCI move string length: %s", moveStr)
	}

	fromFile := int(moveStr[0] - 'a' + 1)
	fromRank := int(moveStr[1] - '0')
	toFile := int(moveStr[2] - 'a' + 1)
	toRank := int(moveStr[3] - '0')

	fromLoc := Location{
		File: fromFile,
		Rank: fromRank,
	}

	toLoc := Location{
		File: toFile,
		Rank: toRank,
	}

	if !fromLoc.IsValid() || !toLoc.IsValid() {
		return fmt.Errorf("parsed out-of-bounds locations from move: %s", moveStr)
	}

	var activeBoard *BoardState
	var opponentBoard *BoardState

	if m.ActiveColor == White {
		activeBoard = m.WhitePlayer.Board
		opponentBoard = m.BlackPlayer.Board
	} else {
		activeBoard = m.BlackPlayer.Board
		opponentBoard = m.WhitePlayer.Board
	}

	movingPiece, exists := (*activeBoard)[fromLoc]
	if !exists {
		return fmt.Errorf("no piece found at source location %s", moveStr)
	}

	//EP Logic
	if movingPiece.Type == Pawn && toLoc.File != fromLoc.File {
		_, targetOccupied := (*opponentBoard)[toLoc]
		if !targetOccupied {
			enemyPawnRank := fromLoc.Rank
			enemyPawnLoc := Location{
				File: toLoc.File,
				Rank: enemyPawnRank,
			}
			delete(*opponentBoard, enemyPawnLoc)
			m.HalfMoveClock = 0
		}
	}

	//capture - this logic needs to come after the ep logic
	_, exists = (*opponentBoard)[toLoc]
	if exists {
		m.HalfMoveClock = 0
		delete(*opponentBoard, toLoc)
	}

	//Castling Logic
	if movingPiece.Type == King && abs(toLoc.File-fromLoc.File) == 2 {
		if toLoc.File == 7 {
			delete(*activeBoard, Location{File: 8, Rank: toLoc.Rank})
			(*activeBoard)[Location{File: 6, Rank: toLoc.Rank}] = Piece{Type: Rook, Color: m.ActiveColor}
		}
		if toLoc.File == 3 {
			delete(*activeBoard, Location{File: 1, Rank: toLoc.Rank})
			(*activeBoard)[Location{File: 6, Rank: toLoc.Rank}] = Piece{Type: Rook, Color: m.ActiveColor}
		}
	}

	//Promotion Logic
	if movingPiece.Type == Pawn && (toLoc.Rank == 8 || toLoc.Rank == 1) {
		if len(moveStr) == 5 {
			promoChar := moveStr[4]
			switch promoChar {
			case 'n':
				movingPiece.Type = Knight
			case 'b':
				movingPiece.Type = Bishop
			case 'r':
				movingPiece.Type = Rook
			case 'q':
				movingPiece.Type = Queen
			}
		} else {
			return fmt.Errorf("No promotion piece specified: %s", moveStr)
		}
	}

	//Update state flags - this might have to come at the very start
	m.updateFlagsAndClocks(movingPiece, fromLoc, toLoc)

	//Complete the move
	delete(*activeBoard, fromLoc)
	(*activeBoard)[toLoc] = movingPiece
	if m.ActiveColor == White {
		m.ActiveColor = Black
	} else {
		m.ActiveColor = White
		m.FullMoveNumber++
	}

	return nil
}

func (m *MatchState) updateFlagsAndClocks(piece Piece, from, to Location) {
	combined := ConcatenateBoardState(m.WhitePlayer.Board, m.BlackPlayer.Board)
	_, isCapture := (*combined)[to]

	if piece.Type == Pawn || isCapture {
		m.HalfMoveClock = 0
	} else {
		m.HalfMoveClock++
	}

	//Set EP target
	if piece.Type == Pawn && abs(to.Rank-from.Rank) == 2 {
		middleRank := (from.Rank + to.Rank) / 2
		fileChar := rune('a' + from.File - 1)
		m.EnPassantTarget = fmt.Sprintf("%c%d", fileChar, middleRank)
	} else {
		m.EnPassantTarget = "-"
	}

	//Castling Rights
	if piece.Type == King {
		if piece.Color == White {
			m.CastlingRights = removeRights(m.CastlingRights, "KQ")
		} else {
			m.CastlingRights = removeRights(m.CastlingRights, "kq")
		}
	}

	if piece.Type == Rook {
		if piece.Color == White {
			if from.File == 1 && from.Rank == 1 {
				m.CastlingRights = removeRights(m.CastlingRights, "Q")
			}
			if from.File == 8 && from.Rank == 1 {
				m.CastlingRights = removeRights(m.CastlingRights, "K")
			}
		} else {
			if from.File == 1 && from.Rank == 8 {
				m.CastlingRights = removeRights(m.CastlingRights, "q")
			}
			if from.File == 8 && from.Rank == 8 {
				m.CastlingRights = removeRights(m.CastlingRights, "k")
			}
		}
	}
}

func removeRights(current, toRemove string) string {
	for _, char := range toRemove {
		current = strings.ReplaceAll(current, string(char), "")
	}
	if current == "" {
		return "-"
	}
	return current
}

func abs(num int) int {
	if num < 0 {
		return -num
	}
	return num
}
