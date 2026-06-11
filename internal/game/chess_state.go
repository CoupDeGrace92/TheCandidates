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
	ID    string     `json:"id"`
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

	return fmt.Sprintf("%s %s %s %s %d %d",
		fenBoard,
		m.ActiveColor,
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
