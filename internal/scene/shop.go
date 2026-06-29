package scene

import (
	"fmt"
	"image/color"
	"time"

	"github.com/CoupDeGrace92/candidates/internal/draft"
	"github.com/CoupDeGrace92/candidates/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type ShopScene struct {
	profile       *game.PlayerProfile
	manager       *draft.DraftManager
	tray          draft.ShopTray
	statusMessage string
	gameOver      bool

	// Runtime calculated boundary areas for scaling calculations
	lastWinW   float64
	lastWinH   float64
	boardX     float64
	boardY     float64
	boardSize  float64
	squareSize float64

	//Temp reroll button with collision boundaries
	rerollBtn imageRect

	//Input Tracking
	selectedBenchIndex  int           //set to -1 if no bench piece is selected
	selectedBoardSquare game.Location //set to (0,0) if no board tile is selected
	isBenchSelected     bool
	isBoardSelected     bool

	lastClickTime   time.Time     //Used for double clicks
	lastClickSquare game.Location //check for continuity on double clicks - both clicks must be same square
}

type imageRect struct {
	x, y, w, h float64
}

func (r imageRect) Contains(x, y int) bool {
	fx, fy := float64(x), float64(y)
	return fx >= r.x && fx <= r.x+r.w && fy >= r.y && fy <= r.y+r.h
}

func NewShopScene(profile *game.PlayerProfile, manager *draft.DraftManager) *ShopScene {
	startingTray := manager.GenerateFreshTray(profile.BoardAndBench.Squares)

	return &ShopScene{
		profile:       profile,
		manager:       manager,
		tray:          startingTray,
		statusMessage: "Welcome to the Draft Phase! Select items to purchase.",
		rerollBtn:     imageRect{x: 480, y: 560, w: 120, h: 40},
	}
}

func (s *ShopScene) Update() error {
	// Catch mouse inputs strictly on the single frame the click occurs
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		now := time.Now()
		bb := s.profile.BoardAndBench

		// ==================================================
		// 1. EVALUATE SHOP BUTTONS AND CARD PURCHASES FIRST
		// ==================================================
		if s.rerollBtn.Contains(mx, my) {
			_ = s.manager.ProcessReroll(&s.tray, s.profile)
			s.clearSelection()
			return nil
		}
		for _, card := range s.tray.Units {
			if s.getUnitItemRect(card.ID).Contains(mx, my) {
				_ = s.manager.BuyItem(&s.tray, card.ID, s.profile)
				return nil
			}
		}

		// ==================================================
		// 2. INTERCEPT CHESSBOARD INTERACTION BOXES
		// ==================================================
		for file := 1; file <= 8; file++ {
			for rank := 1; rank <= 8; rank++ {
				x := s.boardX + float64(file-1)*s.squareSize
				y := s.boardY + float64(8-rank)*s.squareSize
				squareRect := imageRect{x: x, y: y, w: s.squareSize, h: s.squareSize}

				if squareRect.Contains(mx, my) {
					currentLoc := game.Location{File: file, Rank: rank}
					_, pieceOccupied := (*bb.Board)[currentLoc]

					// ----------------------------------------------
					// BOUNDARY A: DOUBLE-CLICK TO RECALL PIECE TO BENCH
					// ----------------------------------------------
					// Double click threshold around 250-300ms
					if pieceOccupied && now.Sub(s.lastClickTime) < 280*time.Millisecond && s.lastClickSquare == currentLoc {
						if bb.BoardToBench(currentLoc) {
							s.statusMessage = "Piece recalled cleanly back to your Bench."
							s.clearSelection()
							return nil
						}
					}

					s.lastClickTime = now
					s.lastClickSquare = currentLoc

					// ----------------------------------------------
					// BOUNDARY B: COMPLETE ACTIVE SHUFFLING SELECTION
					// ----------------------------------------------
					if s.isBenchSelected {
						if bb.BenchToBoard(s.selectedBenchIndex, currentLoc) {
							s.statusMessage = "Unit deployed successfully onto the board grid."
						} else {
							s.statusMessage = "Placement Failed: Square is locked or occupied."
						}
						s.clearSelection()
						return nil
					}

					if s.isBoardSelected {
						if bb.BoardToBoard(s.selectedBoardSquare, currentLoc) {
							s.statusMessage = "Unit re-positioned on the board."
						} else {
							s.statusMessage = "Movement Failed: Target square is invalid."
						}
						s.clearSelection()
						return nil
					}

					// ----------------------------------------------
					// BOUNDARY C: INITIALIZE FRESH SELECTION SOURCE
					// ----------------------------------------------
					// If the square is an active card offer in the shop, purchase it instead of selecting
					isOffer := false
					for _, card := range s.tray.Squares {
						if card.UnlockSquare == currentLoc {
							_ = s.manager.BuyItem(&s.tray, card.ID, s.profile)
							isOffer = true
							break
						}
					}

					if !isOffer && pieceOccupied {
						s.selectedBoardSquare = currentLoc
						s.isBoardSelected = true
						s.statusMessage = fmt.Sprintf("Selected piece at %c%d. Click destination tile, or double-click to recall.", rune('a'+file-1), rank)
					}
					return nil
				}
			}
		}

		// ==================================================
		// 3. INTERCEPT BENCH SELECTION TRACKING BOXES
		// ==================================================
		benchY := s.boardY + s.boardSize + 15
		slotPadding := 8.0
		slotSize := s.lastWinH*0.18 - 45
		startX := 20.0
		currentY := benchY + 35

		for idx := range bb.Bench {
			slotX := startX + float64(idx)*(slotSize+slotPadding)
			if slotX+slotSize > s.lastWinW-20 {
				slotX = startX + float64(idx%6)*(slotSize+slotPadding)
				currentY = benchY + 35 + slotSize + slotPadding
			}

			benchRect := imageRect{x: slotX, y: currentY, w: slotSize, h: slotSize}
			if benchRect.Contains(mx, my) {
				s.clearSelection()
				s.selectedBenchIndex = idx
				s.isBenchSelected = true
				s.statusMessage = fmt.Sprintf("Bench unit selected. Click any highlighted territory square on the board to deploy.")
				return nil
			}
		}

		// Clicked empty background wallpaper: Clear active placement buffers cleanly
		s.clearSelection()
	}
	return nil
}

func (s *ShopScene) clearSelection() {
	s.selectedBenchIndex = -1
	s.selectedBoardSquare = game.Location{File: 0, Rank: 0}
	s.isBenchSelected = false
	s.isBoardSelected = false
}

func (s *ShopScene) Draw(screen *ebiten.Image) {
	screen.Fill(bColor(30, 30, 35))

	bounds := screen.Bounds()
	winW := float64(bounds.Dx())
	winH := float64(bounds.Dy())

	s.lastWinW = winW
	s.lastWinH = winH

	// ==========================================
	// 1. DYNAMIC RESPONSIVE LAYOUT CALCULATOR
	// ==========================================
	// Allocate proportions: Top Shop Tray takes 18%, Bottom Bench takes 18%
	shopTrayH := winH * 0.18
	benchTrayH := winH * 0.18
	usableBoardH := winH - shopTrayH - benchTrayH - 60 // Leave 60px padding for status text

	s.squareSize = usableBoardH / 8
	if (winW-100)/8 < s.squareSize {
		s.squareSize = (winW - 100) / 8
	}
	s.boardSize = s.squareSize * 8

	s.boardX = (winW - s.boardSize) / 2
	s.boardY = shopTrayH + 20

	// ==========================================
	// 2. RENDER TOP ROW SHOP itemS (UNITS)
	// ==========================================
	ebitenutil.DrawRect(screen, 0, 0, winW, shopTrayH, bColor(45, 45, 50))
	headerTextH := shopTrayH * 0.15
	s.DrawScaledText(screen, fmt.Sprintf("GOLD: %d G | SKILL: %d / 20", s.profile.Gold, s.profile.SkillLevel), 20, 10, headerTextH, color.White)

	for _, item := range s.tray.Units {
		r := s.getUnitItemRect(item.ID)
		ebitenutil.DrawRect(screen, r.x, r.y, r.w, r.h, bColor(60, 60, 75))

		sprite := GetPieceSprite(game.Piece{Type: item.PieceType, Color: game.White})
		if sprite != nil {
			op := &ebiten.DrawImageOptions{}

			targetSpriteH := r.h * .6
			scaleX := targetSpriteH / float64(SpriteW)
			scaleY := targetSpriteH / float64(SpriteH)

			scaledPieceW := float64(SpriteW) * scaleX
			scaledPieceH := float64(SpriteH) * scaleY

			paddingX := (r.w - scaledPieceW) / 2.0
			paddingY := (r.h - scaledPieceH) / 3.0

			op.GeoM.Scale(scaleX, scaleY)
			op.GeoM.Translate(r.x+paddingX, r.y+paddingY)
			screen.DrawImage(sprite, op)
		}

	}

	//Reroll button proportions
	btnW := s.squareSize * 2.2
	btnH := s.squareSize * 0.7
	btnY := (shopTrayH / 2.0) - (btnH / 2.0)

	s.rerollBtn = imageRect{x: winW - btnW - 20, y: btnY, w: btnW, h: btnH}
	ebitenutil.DrawRect(screen, s.rerollBtn.x, s.rerollBtn.y, s.rerollBtn.w, s.rerollBtn.h, bColor(140, 110, 40))
	s.DrawScaledText(screen, "REROLL (1G)", s.rerollBtn.x+15, s.rerollBtn.y+s.rerollBtn.h/2-6, s.rerollBtn.h*0.2, color.White)

	// ==========================================
	// 3. RENDER BOARD
	// ==========================================
	bb := s.profile.BoardAndBench

	for rank := 8; rank >= 1; rank-- {
		for file := 1; file <= 8; file++ {
			x := s.boardX + float64(file-1)*s.squareSize
			y := s.boardY + float64(8-rank)*s.squareSize

			loc := game.Location{File: file, Rank: rank}
			_, isOwned := s.profile.BoardAndBench.Squares[loc]

			var matchingOffer *draft.ShopItem
			for i, item := range s.tray.Squares {
				if item.UnlockSquare == loc {
					matchingOffer = &s.tray.Squares[i]
					break
				}
			}

			var tileColor color.RGBA
			if isOwned {
				tileColor = s.profile.Theme.LightSquare
				if (rank+file)%2 == 0 {
					tileColor = s.profile.Theme.DarkSquare
				}
			} else {
				tileColor = color.RGBA{60, 60, 65, 255}
				if (rank+file)%2 == 0 {
					tileColor = color.RGBA{45, 45, 50, 255}
				}
			}

			ebitenutil.DrawRect(screen, x, y, s.squareSize-1, s.squareSize-1, tileColor)

			isSelectedSquare := s.isBoardSelected && s.selectedBoardSquare == loc
			if isSelectedSquare {
				// Draws a semi-transparent glowing gold backdrop layer directly inside the tile bounds
				ebitenutil.DrawRect(screen, x, y, s.squareSize-1, s.squareSize-1, color.RGBA{240, 200, 50, 120})
			}

			if matchingOffer != nil {
				ebitenutil.DrawRect(screen, x, y, s.squareSize-1, 3, color.RGBA{210, 160, 50, 255})
				ebitenutil.DrawRect(screen, x, y, 3, s.squareSize-1, color.RGBA{210, 160, 50, 255})

				priceTag := fmt.Sprintf("$%d", matchingOffer.Cost)
				s.DrawScaledText(screen, priceTag, x+s.squareSize/2-10, y+s.squareSize/2-6, s.squareSize*0.3, color.RGBA{210, 160, 50, 255})
			}

			//Render the pieces on the board
			if piece, occupied := (*bb.Board)[loc]; occupied {
				if sprite := GetPieceSprite(piece); sprite != nil {
					op := &ebiten.DrawImageOptions{}

					scaleX := s.squareSize / float64(SpriteW)
					scaleY := s.squareSize / float64(SpriteH)

					op.GeoM.Scale(scaleX, scaleY)
					op.GeoM.Translate(x, y)

					screen.DrawImage(sprite, op)
				}
			}
		}
	}

	// ==========================================
	// 4. RENDER EXPANDED BOTTOM ROW (BENCH INVENTORY)
	// ==========================================
	benchY := s.boardY + s.boardSize + 15
	ebitenutil.DrawRect(screen, 0, benchY, winW, benchTrayH, bColor(40, 40, 45))

	benchTextH := benchTrayH * 0.12
	s.DrawScaledText(screen, fmt.Sprintf("BENCH INVENTORY (%d items):", len(s.profile.BoardAndBench.Bench)), 20, benchY+10, benchTextH, color.White)

	slotPadding := 8.0
	slotSize := benchTrayH - 45
	startX := 20.0
	currentY := benchY + 35

	for idx, piece := range s.profile.BoardAndBench.Bench {
		slotX := startX + float64(idx)*(slotSize+slotPadding)

		if slotX+slotSize > winW-20 {
			slotX = startX + float64(idx%6)*(slotSize+slotPadding)
			currentY = benchY + 35 + slotSize + slotPadding
		}

		isThisBenchSelected := s.isBenchSelected && s.selectedBenchIndex == idx
		if isThisBenchSelected {
			borderGlow := 3.0
			ebitenutil.DrawRect(screen, slotX-borderGlow, currentY-borderGlow, slotSize+(borderGlow*2), slotSize+(borderGlow*2), color.RGBA{0, 255, 230, 255})
		}

		sprite := GetPieceSprite(piece)
		if sprite != nil {
			op := &ebiten.DrawImageOptions{}

			targetSpriteH := slotSize * .75
			scaleX := targetSpriteH / float64(SpriteW)
			scaleY := targetSpriteH / float64(SpriteH)

			scaledPieceW := float64(SpriteW) * scaleX
			scaledPieceH := float64(SpriteH) * scaleY

			paddingX := (slotSize - scaledPieceW) / 2.0
			paddingY := (slotSize - scaledPieceH) / 3.0

			op.GeoM.Scale(scaleX, scaleY)
			op.GeoM.Translate(slotX+paddingX, currentY+paddingY)

			screen.DrawImage(sprite, op)
		}
	}

	// 5. Status Footer Layout
	footerTextH := winH * .022
	footerY := winH - footerTextH - 10
	s.DrawScaledText(screen, fmt.Sprintf("System Log: %s", s.statusMessage), 20, footerY, footerTextH, color.RGBA{180, 180, 180, 255})
}

func (s *ShopScene) getUnitItemRect(itemID string) imageRect {
	for i, c := range s.tray.Units {
		if c.ID == itemID {
			itemW := (s.lastWinW - 200) / 4
			itemH := s.lastWinH * 0.12
			x := 40.0 + float64(i)*(itemW+15.0)
			y := (s.lastWinH * 0.18 / 2) - (itemH / 2) + 10

			return imageRect{x: x, y: y, w: itemW, h: itemH}
		}
	}
	return imageRect{}
}

func (s *ShopScene) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (s *ShopScene) Destroy() {
	//Placeholder for between screens
	fmt.Println("Shop components flushed.")
}

func bColor(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 255} }

func (s *ShopScene) DrawScaledText(screen *ebiten.Image, text string, x, y, targetHeight float64, clr color.Color) {
	textW := len(text) * 7
	if textW == 0 {
		textW = 1
	}
	textBuffer := ebiten.NewImage(textW, 14)

	ebitenutil.DebugPrint(textBuffer, text)

	scaleFactor := targetHeight / 12.0

	op := &ebiten.DrawImageOptions{}

	if clr != nil {
		r, g, b, a := clr.RGBA()
		op.ColorScale.Scale(float32(r)/65535, float32(g)/65535, float32(b)/65535, float32(a)/65535)
	}

	op.GeoM.Scale(scaleFactor, scaleFactor)
	op.GeoM.Translate(x, y)

	screen.DrawImage(textBuffer, op)
}
