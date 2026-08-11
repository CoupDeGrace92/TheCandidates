package game

import "image/color"

type BoardTheme struct {
	LightSquare color.RGBA
	DarkSquare  color.RGBA
}

type PlayerProfile struct {
	PlayerID     string     `json:"player_id"`
	IsHuman      bool       `json:"is_human"`
	Gold         int        `json:"gold"`
	BotTotalGold int        `json:"bot_total_gold"`
	Color        PieceColor `json:"color"`

	BotStaticCatalog []interface{} `json:"static_catalog"` //Note: []interface{} to get around circular package dependencies

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

	profile.Color = newColor
	bb := profile.BoardAndBench

	for i := range bb.Bench {
		bb.Bench[i].Color = newColor
	}

	tPieces := make(map[Location]Piece)
	for loc, piece := range *bb.Board {
		piece.Color = newColor
		newLoc := loc.Transform()
		tPieces[newLoc] = piece
	}
	(*bb.Board) = tPieces

	tSquares := make(map[Location]struct{})
	for loc := range bb.Squares {
		tSquares[loc.Transform()] = struct{}{}
	}
	bb.Squares = tSquares
}
