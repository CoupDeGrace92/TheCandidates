package game

import "image/color"

type BoardTheme struct {
	LightSquare color.RGBA
	DarkSquare  color.RGBA
}

type PlayerProfile struct {
	PlayerID string     `json:"player_id"`
	IsHuman  bool       `json:"is_human"`
	Gold     int        `json:"gold"`
	Color    PieceColor `json:"color"`

	SkillLevel      int `json:"skill_level"`
	MoveTimeMs      int `json:"move_time_ms"`
	MaxDrawishTurns int `json:"max_drawish_turns"`

	SpriteSheetPath string     `json:"sprite_sheet_path"`
	Theme           BoardTheme `json:"theme"`

	BoardAndBench *PlayerPieces `json:"board_and_bench"`
}

func NewDefaultProfile(id string, isHuman bool) *PlayerProfile {
	darkTile := color.RGBA{120, 135, 120, 255}
	lightTile := color.RGBA{235, 235, 235, 235}

	allowed := make(map[Location]struct{})

	for rank := 1; rank <= 2; rank++ {
		for file := 1; file <= 8; file++ {
			allowed[Location{File: file, Rank: rank}] = struct{}{}
		}
	}
	King := Piece{
		Type:  King,
		Color: White,
	}
	wbs := make(BoardState)
	bb := PlayerPieces{
		Board:   &wbs,
		Bench:   []Piece{King},
		Squares: allowed,
	}

	return &PlayerProfile{
		PlayerID:        id,
		IsHuman:         isHuman,
		Gold:            10,
		Color:           White,
		SkillLevel:      7,
		MoveTimeMs:      150,
		MaxDrawishTurns: 14,
		SpriteSheetPath: "assets/images/GenericChessPiecesSprite.png",
		Theme: BoardTheme{
			LightSquare: lightTile,
			DarkSquare:  darkTile,
		},
		BoardAndBench: &bb,
	}
}

func (profile *PlayerProfile) SwitchColorIfDifferent(newColor PieceColor) {
	if profile == nil || profile.Color == newColor || profile.BoardAndBench == nil {
		return
	}

	bb := profile.BoardAndBench

	newSquares := make(map[Location]struct{})
	for loc := range bb.Squares {
		nLoc := loc.Transform()
		newSquares[nLoc] = struct{}{}
	}
	bb.Squares = newSquares

	newBoard := make(BoardState)
	for loc, piece := range *bb.Board {
		piece.Color = newColor
		newBoard[loc.Transform()] = piece
	}
	*bb.Board = newBoard

	for i := range bb.Bench {
		bb.Bench[i].Color = newColor
	}

	profile.Color = newColor
}
