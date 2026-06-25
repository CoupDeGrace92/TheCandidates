package scene

import (
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/CoupDeGrace92/candidates/internal/game"
)

var pieceCache map[string]*ebiten.Image

var SpriteW int
var SpriteH int

func LoadAssets(path string) {
	masterSheet, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		log.Fatalf("Fatal: failed to load pieces spritesheet asset: %v", err)
	}

	pieceCache = make(map[string]*ebiten.Image)

	bounds := masterSheet.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	SpriteW = imgWidth / 6
	SpriteH = imgHeight / 2

	pieceTypes := []game.PieceType{
		game.King,
		game.Queen,
		game.Bishop,
		game.Knight,
		game.Rook,
		game.Pawn,
	}
	colors := []game.PieceColor{
		game.White,
		game.Black,
	}

	for row, colorVal := range colors {
		for col, typeVal := range pieceTypes {
			startX := col * SpriteW
			startY := row * SpriteH
			endX := startX + SpriteW
			endY := startY + SpriteH

			subImg := masterSheet.SubImage(image.Rect(startX, startY, endX, endY))

			ebitenSprite := subImg.(*ebiten.Image)

			cacheKey := string(colorVal) + "_" + string(typeVal)
			pieceCache[cacheKey] = ebitenSprite
		}
	}
	if SpriteW == 0 || SpriteH == 0 {
		log.Fatalf("Fatal: Slices evaluation resulted in zero-size bounds. Image width: %d", bounds.Dx())
	}
	log.Printf("Cached chess piece sprites")
}

func GetPieceSprite(p game.Piece) *ebiten.Image {
	if pieceCache == nil {
		return nil
	}
	cacheKey := string(p.Color) + "_" + string(p.Type)
	return pieceCache[cacheKey]
}
