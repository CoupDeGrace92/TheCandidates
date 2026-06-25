package game

import "image/color"

type BoardTheme struct {
	LightSquare color.RGBA
	DarkSquare  color.RGBA
}

type PlayerProfile struct {
	PlayerID string `json:"player_id"`
	IsHuman  bool   `json:"is_human"`
	Gold     int    `json:"gold"`

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

	bb := PlayerPieces{
		Board: &BoardState{},
		Bench: []Piece{},
	}

	return &PlayerProfile{
		PlayerID:        id,
		IsHuman:         isHuman,
		Gold:            10,
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
