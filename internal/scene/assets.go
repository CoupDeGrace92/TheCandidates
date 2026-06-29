package scene

import (
	"errors"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/CoupDeGrace92/candidates/internal/game"
)

var masterSheet *ebiten.Image
var pieceCache map[string]*ebiten.Image
var preScaledCache map[string]*ebiten.Image

var SpriteW int
var SpriteH int

func LoadAssets(path string) {
	var err error
	masterSheet, _, err = ebitenutil.NewImageFromFile(path)
	if err != nil {
		log.Fatalf("Fatal: failed to load pieces spritesheet asset: %v", err)
	}

	pieceCache = make(map[string]*ebiten.Image)
	preScaledCache = make(map[string]*ebiten.Image)

	bounds := masterSheet.Bounds()

	SpriteW = bounds.Dx() / 6
	SpriteH = bounds.Dy() / 2

	//If we have sprite sheets with different orders, we have to change this or inject dynamically
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

func RegenerateScaledUICache(targetSize float64) error {
	if preScaledCache == nil || masterSheet == nil {
		return errors.New("No master sprite sheet or prescaled cache detected")
	}
	if targetSize < 5 {
		targetSize = 60
	}
	intSize := int(targetSize)
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

	for _, colorVal := range colors {
		for _, typeVal := range pieceTypes {
			cacheKey := string(colorVal) + "_" + string(typeVal)
			rawSprite := pieceCache[cacheKey]
			if rawSprite == nil {
				continue
			}

			scaledImg := ebiten.NewImage(intSize, intSize)

			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterNearest
			op.GeoM.Scale(targetSize/float64(SpriteW), targetSize/float64(SpriteH))
			scaledImg.DrawImage(rawSprite, op)

			preScaledCache[cacheKey] = scaledImg
		}
	}

	return nil
}

func GetPieceSprite(p game.Piece) *ebiten.Image {
	if pieceCache == nil {
		return nil
	}
	cacheKey := string(p.Color) + "_" + string(p.Type)
	return pieceCache[cacheKey]
}

func GetPreScaledSprite(p game.Piece) *ebiten.Image {
	if preScaledCache == nil {
		return nil
	}
	cacheKey := string(p.Color) + "_" + string(p.Type)
	return preScaledCache[cacheKey]
}
