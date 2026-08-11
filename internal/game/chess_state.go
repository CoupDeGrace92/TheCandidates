package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"unicode"
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

type PlayerPieces struct {
	//This profile might move - this also might indicate which assets to pull for the player
	Board   *BoardState           `json:"board"`
	Bench   []Piece               `json:"bench"`
	Squares map[Location]struct{} `json:"squares"`
}

func NewPlayerPieces() *PlayerPieces {
	allowed := make(map[Location]struct{})
	for rank := 1; rank <= 2; rank++ {
		for file := 1; file <= 8; file++ {
			allowed[Location{File: file, Rank: rank}] = struct{}{}
		}
	}

	bs := make(BoardState)
	return &PlayerPieces{
		Board:   &bs,
		Bench:   []Piece{},
		Squares: allowed,
	}
}

type MatchState struct {
	Board BoardState `json:"board"`

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

	var rows []string
	// FEN strings serialize the grid starting from Top Rank (8) down to Bottom Rank (1)
	for rank := 8; rank >= 1; rank-- {
		var rankStringBuilder strings.Builder
		emptyCount := 0

		for file := 1; file <= 8; file++ {
			loc := Location{Rank: rank, File: file}

			if piece, occupied := m.Board[loc]; occupied {
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
		m.CastlingRights,
		ep,
		m.HalfMoveClock,
		m.FullMoveNumber,
	)
}

func (m *MatchState) InitializeCastlingRights() {
	rights := ""

	if wk, ok := m.Board[Location{Rank: 1, File: 5}]; ok && wk.Type == King && wk.Color == White {
		if r, ok := m.Board[Location{Rank: 1, File: 8}]; ok && r.Type == Rook && r.Color == White {
			rights += "K"
		}
		if r, ok := m.Board[Location{Rank: 1, File: 1}]; ok && r.Type == Rook && r.Color == White {
			rights += "Q"
		}
	}

	if bk, ok := m.Board[Location{Rank: 8, File: 5}]; ok && bk.Type == King && bk.Color == Black {
		if r, ok := m.Board[Location{Rank: 8, File: 8}]; ok && r.Type == Rook && r.Color == Black {
			rights += "k"
		}
		if r, ok := m.Board[Location{Rank: 8, File: 1}]; ok && r.Type == Rook && r.Color == Black {
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

	// Parse UCI move string characters into clean 1-indexed integers
	fromFile := int(moveStr[0] - 'a' + 1)
	fromRank := int(moveStr[1] - '0')
	toFile := int(moveStr[2] - 'a' + 1)
	toRank := int(moveStr[3] - '0')

	fromLoc := Location{File: fromFile, Rank: fromRank}
	toLoc := Location{File: toFile, Rank: toRank}

	if !fromLoc.IsValid() || !toLoc.IsValid() {
		return fmt.Errorf("parsed out-of-bounds locations from move: %s", moveStr)
	}

	// FIXED: Extract the moving piece directly out of our unified Board container,
	// dropping the broken cross-profile pointer routing completely!
	movingPiece, exists := m.Board[fromLoc]
	if !exists {
		return fmt.Errorf("no piece found at source location %s", moveStr)
	}

	// 1. En Passant Capture Logic
	if movingPiece.Type == Pawn && toLoc.File != fromLoc.File {
		if _, targetOccupied := m.Board[toLoc]; !targetOccupied {
			// Capturing a pawn via En Passant removes the target pawn from the rank behind it
			delete(m.Board, Location{File: toLoc.File, Rank: fromLoc.Rank})
			m.HalfMoveClock = 0
		}
	}

	// 2. Standard Capture Logic
	if _, targetOccupied := m.Board[toLoc]; targetOccupied {
		m.HalfMoveClock = 0
		delete(m.Board, toLoc) // Remove the captured piece from the unified combat grid
	}

	// 3. Castling Secondary Rook Shifting Logic
	if movingPiece.Type == King && abs(toLoc.File-fromLoc.File) == 2 {
		if toLoc.File == 7 { // Kingside
			delete(m.Board, Location{File: 8, Rank: toLoc.Rank})
			m.Board[Location{File: 6, Rank: toLoc.Rank}] = Piece{Type: Rook, Color: m.ActiveColor}
		} else if toLoc.File == 3 { // Queenside
			delete(m.Board, Location{File: 1, Rank: toLoc.Rank})
			m.Board[Location{File: 4, Rank: toLoc.Rank}] = Piece{Type: Rook, Color: m.ActiveColor}
		}
	}

	// 4. Pawn Promotion Logic
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
			return fmt.Errorf("no promotion piece specified: %s", moveStr)
		}
	}

	// Update system clocks and En Passant flags on the unified state
	m.updateFlagsAndClocks(movingPiece, fromLoc, toLoc)

	// 5. Complete the Move State Mutation
	delete(m.Board, fromLoc)
	m.Board[toLoc] = movingPiece

	// Shift active color turns and advance move counters
	if m.ActiveColor == White {
		m.ActiveColor = Black
	} else {
		m.ActiveColor = White
		m.FullMoveNumber++
	}

	return nil
}

func (m *MatchState) updateFlagsAndClocks(piece Piece, from, to Location) {
	_, isCapture := m.Board[to]

	if piece.Type == Pawn || isCapture {
		m.HalfMoveClock = 0
	} else {
		m.HalfMoveClock++
	}

	// Set En Passant target string for the next turn loop
	if piece.Type == Pawn && abs(to.Rank-from.Rank) == 2 {
		middleRank := (from.Rank + to.Rank) / 2
		fileChar := rune('a' + from.File - 1)
		m.EnPassantTarget = fmt.Sprintf("%c%d", fileChar, middleRank)
	} else {
		m.EnPassantTarget = "-"
	}

	// Re-evaluate remaining Castling Rights flags
	if piece.Type == King {
		if piece.Color == White {
			m.CastlingRights = removeRights(m.CastlingRights, "KQ")
		} else {
			m.CastlingRights = removeRights(m.CastlingRights, "kq")
		}
	} else if piece.Type == Rook {
		if from.Rank == 1 && from.File == 8 {
			m.CastlingRights = removeRights(m.CastlingRights, "K")
		}
		if from.Rank == 1 && from.File == 1 {
			m.CastlingRights = removeRights(m.CastlingRights, "Q")
		}
		if from.Rank == 8 && from.File == 8 {
			m.CastlingRights = removeRights(m.CastlingRights, "k")
		}
		if from.Rank == 8 && from.File == 1 {
			m.CastlingRights = removeRights(m.CastlingRights, "q")
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
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

type PositionTracker map[string]int

func (pt PositionTracker) RecordPosition(fullFEN string) bool {
	parts := strings.Split(fullFEN, " ")
	if len(parts) < 4 {
		return false
	}

	positionKey := strings.Join(parts[:4], " ")
	pt[positionKey]++

	return pt[positionKey] >= 3
}

func (p *PlayerPieces) BenchToBoard(idx int, sq Location) bool {
	if idx < 0 || idx >= len(p.Bench) {
		return false
	}

	if _, allowed := p.Squares[sq]; !allowed {
		return false
	}
	if _, ok := (*p.Board)[sq]; ok {
		return false
	}

	piece := p.Bench[idx]
	p.Bench = append(p.Bench[:idx], p.Bench[idx+1:]...)
	(*p.Board)[sq] = piece
	return true
}

func (p *PlayerPieces) BoardToBench(sq Location) bool {
	if piece, ok := (*p.Board)[sq]; ok {
		p.Bench = append(p.Bench, piece)
		delete(*p.Board, sq)
		return true
	}
	return false
}

func (p *PlayerPieces) BoardToBoard(init, target Location) bool {
	if _, allowed := p.Squares[target]; !allowed {
		return false
	}
	if _, ok := (*p.Board)[target]; ok {
		return false
	}
	if piece, ok := (*p.Board)[init]; ok {
		(*p.Board)[target] = piece
		delete(*p.Board, init)
		return true
	}
	return false
}

func (p *PlayerPieces) RandomOpenSquare(excluded map[Location]struct{}) (Location, bool) {
	var eligible []Location
	for loc := range p.Squares {
		if excluded != nil {
			if _, isExcluded := excluded[loc]; isExcluded {
				continue
			}
		}
		eligible = append(eligible, loc)
	}
	if len(eligible) == 0 {
		return Location{}, false
	}

	chosenIndex := rand.Intn(len(eligible))
	return eligible[chosenIndex], true
}

func (l Location) ToRelative(playerColor PieceColor) Location {
	if playerColor == Black {
		return l.Transform()
	}
	return l
}

func (l Location) Transform() Location {
	return Location{
		File: 9 - l.File,
		Rank: 9 - l.Rank,
	}
}

func ParseFENForColor(fen string, targetColor PieceColor) BoardState {
	board := make(BoardState)
	if fen == "" {
		return board
	}

	parts := strings.Split(fen, " ")
	boardPart := parts[0]
	ranks := strings.Split(boardPart, "/")
	for fenRankIdx, rankStr := range ranks {
		absoluteRank := 8 - fenRankIdx
		file := 1

		for _, char := range rankStr {
			if unicode.IsDigit(char) {
				file += int(char - '0')
			} else {
				pieceColor := White
				if unicode.IsLower(char) {
					pieceColor = Black
				}

				if pieceColor == targetColor {
					var pType PieceType
					switch unicode.ToLower(char) {
					case 'p':
						pType = Pawn
					case 'n':
						pType = Knight
					case 'b':
						pType = Bishop
					case 'r':
						pType = Rook
					case 'q':
						pType = Queen
					case 'k':
						pType = King
					}

					absoluteLoc := Location{File: file, Rank: absoluteRank}

					// If we are extracting Black's side, their pieces naturally live on Ranks 7-8 globally.
					// We must flip their coordinates down via Transform() so they sit on Ranks 1-2 internally,
					// matching how local draft profiles store data.
					var localizedLoc Location
					if targetColor == Black {
						localizedLoc = absoluteLoc.Transform()
					} else {
						localizedLoc = absoluteLoc
					}

					board[localizedLoc] = Piece{
						Type:  pType,
						Color: targetColor,
					}
				}
				file++
			}
		}
	}

	return board
}
