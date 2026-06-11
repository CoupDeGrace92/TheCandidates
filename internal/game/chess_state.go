package game

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

// One thing to note about the ToFEN function - since we only have to convert ToFEN() once, then just pass around FENs
// we may not need to keep track of last move for purposes of e.p. or castling
func (m *MatchState) ToFEN() string {
	return ""
}

func (m *MatchState) InitializeCastlingRights() {
	rights := ""

	if wk, ok := (*m.WhitePlayer.Board)[Location{Rank: 1, File: 5}]; ok && wk.Type == King {
		if r, ok := (*m.WhitePlayer.Board)[Location{Rank: 1, File: 8}]; ok && r.Type == Rook {
			rights += "K"
		}
		if r, ok := (*m.WhitePlayer.Board)[Location{Rank: 1, File: 0}]; ok && r.Type == Rook {
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
