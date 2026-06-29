package scene

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/CoupDeGrace92/candidates/internal/draft"
	"github.com/CoupDeGrace92/candidates/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type DragSource string

const (
	DragFromNone  DragSource = "none"
	DragFromBench DragSource = "bench"
	DragFromBoard DragSource = "board"
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

	lastLayoutH int
	lastLayoutW int

	//Temp reroll button with collision boundaries
	rerollBtn imageRect

	//Input Tracking
	selectedBenchIndex  int           //set to -1 if no bench piece is selected
	selectedBoardSquare game.Location //set to (0,0) if no board tile is selected
	isBenchSelected     bool
	isBoardSelected     bool

	lastClickTime   time.Time     //Used for double clicks
	lastClickSquare game.Location //check for continuity on double clicks - both clicks must be same square

	startX, startY int //Stores initial click pixel - allows us to check drag vs click to click

	dragSource         DragSource
	draggedBenchIdx    int
	draggedBoardSquare game.Location
	isDragging         bool

	dragDrawOp *ebiten.DrawImageOptions
}

type imageRect struct {
	x, y, w, h float64
}

func (r imageRect) Contains(x, y int) bool {
	fx, fy := float64(x), float64(y)
	return fx >= r.x && fx <= r.x+r.w && fy >= r.y && fy <= r.y+r.h
}

func NewShopScene(profile *game.PlayerProfile, manager *draft.DraftManager) *ShopScene {
	startingTray := manager.GenerateFreshTray(profile.BoardAndBench.Squares, profile.Color)

	return &ShopScene{
		profile:       profile,
		manager:       manager,
		tray:          startingTray,
		statusMessage: "Welcome to the Draft Phase! Select items to purchase.",
		rerollBtn:     imageRect{x: 480, y: 560, w: 120, h: 40},
		dragDrawOp:    &ebiten.DrawImageOptions{},
	}
}

func (s *ShopScene) Update() error {
	mx, my := ebiten.CursorPosition()
	bb := s.profile.BoardAndBench
	playerColor := s.profile.Color

	// Helper to find absolute map grid tile coordinates under the current cursor position
	hoveredSquare := game.Location{File: 0, Rank: 0}
	for screenRow := 0; screenRow < 8; screenRow++ {
		for screenCol := 0; screenCol < 8; screenCol++ {
			x := s.boardX + float64(screenCol)*s.squareSize
			y := s.boardY + float64(screenRow)*s.squareSize
			if float64(mx) >= x && float64(mx) < x+s.squareSize && float64(my) >= y && float64(my) < y+s.squareSize {
				if playerColor == game.Black {
					hoveredSquare = game.Location{File: screenCol + 1, Rank: screenRow + 1}
				} else {
					hoveredSquare = game.Location{File: screenCol + 1, Rank: 8 - screenRow}
				}
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		s.startX, s.startY = mx, my // Anchor initial mouse click coordinates

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

		if hoveredSquare.File != 0 {
			if _, occupied := (*bb.Board)[hoveredSquare]; occupied {
				if time.Since(s.lastClickTime) < 280*time.Millisecond && s.lastClickSquare == hoveredSquare {
					if bb.BoardToBench(hoveredSquare) {
						s.statusMessage = "Piece recalled cleanly back to your Bench."
						s.clearSelection()
						return nil
					}
				}
				s.lastClickTime = time.Now()
				s.lastClickSquare = hoveredSquare

				s.dragSource = DragFromBoard
				s.draggedBoardSquare = hoveredSquare
				return nil
			}

			for _, card := range s.tray.Squares {
				if card.UnlockSquare == hoveredSquare {
					_ = s.manager.BuyItem(&s.tray, card.ID, s.profile)
					return nil
				}
			}
		}

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

			if float64(mx) >= slotX && float64(mx) < slotX+slotSize && float64(my) >= currentY && float64(my) < currentY+slotSize {
				s.dragSource = DragFromBench
				s.draggedBenchIdx = idx
				return nil
			}
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && s.dragSource != DragFromNone && !s.isDragging {
		dx := mx - s.startX
		dy := my - s.startY
		distanceMoved := math.Sqrt(float64(dx*dx + dy*dy))

		if distanceMoved > 10.0 {
			s.isDragging = true
			s.statusMessage = "Dragging unit... Drop on a valid highlight square."
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if s.isDragging {
			s.isDragging = false
			if s.dragSource == DragFromBench && hoveredSquare.File != 0 {
				_ = bb.BenchToBoard(s.draggedBenchIdx, hoveredSquare)
			} else if s.dragSource == DragFromBoard {
				if hoveredSquare.File != 0 && hoveredSquare != s.draggedBoardSquare {
					_ = bb.BoardToBoard(s.draggedBoardSquare, hoveredSquare)
				} else if float64(my) >= s.boardY+s.boardSize+15 {
					_ = bb.BoardToBench(s.draggedBoardSquare)
				}
			}
			s.dragSource = DragFromNone
			s.clearSelection()
			return nil
		}

		if s.dragSource != DragFromNone {
			if s.dragSource == DragFromBench {
				s.clearSelection()
				s.selectedBenchIndex = s.draggedBenchIdx
				s.isBenchSelected = true
				s.statusMessage = "Bench unit selected via click. Tap an unlocked board square to deploy."
			} else if s.dragSource == DragFromBoard {
				if s.isBenchSelected && hoveredSquare == s.draggedBoardSquare {
					_ = bb.BenchToBoard(s.selectedBenchIndex, hoveredSquare)
					s.clearSelection()
				} else if s.isBoardSelected && s.selectedBoardSquare != s.draggedBoardSquare {
					_ = bb.BoardToBoard(s.selectedBoardSquare, s.draggedBoardSquare)
					s.clearSelection()
				} else {
					s.clearSelection()
					s.selectedBoardSquare = s.draggedBoardSquare
					s.isBoardSelected = true
					s.statusMessage = "Frontline piece selected via click. Click an open destination tile."
				}
			}
			s.dragSource = DragFromNone
			return nil
		}

		//If clicked on non-piece/non-square, reset clicks
		if hoveredSquare.File != 0 && (s.isBenchSelected || s.isBoardSelected) {
			if s.isBenchSelected {
				_ = bb.BenchToBoard(s.selectedBenchIndex, hoveredSquare)
			} else if s.isBoardSelected {
				_ = bb.BoardToBoard(s.selectedBoardSquare, hoveredSquare)
			}
		}
		s.dragSource = DragFromNone
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

	// =========================================
	//      Core Proportional Allocations
	// =========================================
	shopTrayH := winH * 0.18
	benchTrayH := winH * 0.18
	usableBoardH := winH - shopTrayH - benchTrayH - 60

	s.squareSize = usableBoardH / 8
	if (winW-100)/8 < s.squareSize {
		s.squareSize = (winW - 100) / 8
	}
	s.boardSize = s.squareSize * 8
	s.boardX = (winW - s.boardSize) / 2
	s.boardY = shopTrayH + 20

	bb := s.profile.BoardAndBench
	playerColor := s.profile.Color

	// ==========================================
	// 		        SHOP Renderer
	// ==========================================
	ebitenutil.DrawRect(screen, 0, 0, winW, shopTrayH, bColor(45, 45, 50))

	headerTextH := shopTrayH * 0.15
	s.DrawScaledText(screen, fmt.Sprintf("GOLD: %d G | SPEED: %d / 20", s.profile.Gold, s.profile.SkillLevel), 20, 10, headerTextH, color.White)

	for _, card := range s.tray.Units {
		r := s.getUnitItemRect(card.ID)

		ebitenutil.DrawRect(screen, r.x, r.y, r.w, r.h, bColor(60, 60, 75))

		shopPiece := game.Piece{
			Type:  card.PieceType,
			Color: playerColor,
		}

		if sprite := GetPreScaledSprite(shopPiece); sprite != nil {
			s.dragDrawOp.GeoM.Reset()
			s.dragDrawOp.ColorScale.Reset()

			targetSpriteH := r.h * 0.60
			scaleFactor := targetSpriteH / s.squareSize

			scaledDim := s.squareSize * scaleFactor
			paddingX := (r.w - scaledDim) / 2.0
			paddingY := (r.h - scaledDim) / 3.5

			s.dragDrawOp.GeoM.Scale(scaleFactor, scaleFactor)
			s.dragDrawOp.GeoM.Translate(r.x+paddingX, r.y+paddingY)

			screen.DrawImage(sprite, s.dragDrawOp)
		}

		priceTagStr := fmt.Sprintf("%d G", card.Cost)
		priceTextY := r.y + r.h - (r.h * 0.22)
		s.DrawScaledText(screen, priceTagStr, r.x+15, priceTextY, r.h*0.14, color.White)
	}

	// ==========================================
	// 				Board Renderer
	// ==========================================

	for screenRow := 0; screenRow < 8; screenRow++ {
		for screenCol := 0; screenCol < 8; screenCol++ {
			x := s.boardX + float64(screenCol)*s.squareSize
			y := s.boardY + float64(screenRow)*s.squareSize

			var absoluteLoc game.Location
			if playerColor == game.Black {
				absoluteLoc = game.Location{File: screenCol + 1, Rank: screenRow + 1}
			} else {
				absoluteLoc = game.Location{File: screenCol + 1, Rank: 8 - screenRow}
			}

			_, isOwned := bb.Squares[absoluteLoc]

			var matchingOffer *draft.ShopItem
			for i, card := range s.tray.Squares {
				if card.UnlockSquare == absoluteLoc {
					matchingOffer = &s.tray.Squares[i]
					break
				}
			}

			var tileColor color.RGBA
			if (screenRow+screenCol)%2 == 0 {
				tileColor = s.profile.Theme.LightSquare
			} else {
				tileColor = s.profile.Theme.DarkSquare
			}

			if !isOwned {
				tileColor.R = uint8(float64(tileColor.R) * 0.35)
				tileColor.G = uint8(float64(tileColor.G) * 0.35)
				tileColor.B = uint8(float64(tileColor.B) * 0.35)
			}

			if s.isBoardSelected && s.selectedBoardSquare == absoluteLoc {
				tileColor = color.RGBA{240, 200, 50, 120}
			}

			ebitenutil.DrawRect(screen, x, y, s.squareSize-1, s.squareSize-1, tileColor)

			if matchingOffer != nil {
				ebitenutil.DrawRect(screen, x, y, s.squareSize-1, 3, color.RGBA{210, 160, 50, 255})
				ebitenutil.DrawRect(screen, x, y, 3, s.squareSize-1, color.RGBA{210, 160, 50, 255})

				visualFileChar := rune('a' + screenCol)
				visualRankNum := 8 - screenRow
				if playerColor == game.Black {
					visualRankNum = screenRow + 1
				}
				priceTag := fmt.Sprintf("%c%d\n$%d", visualFileChar, visualRankNum, matchingOffer.Cost)
				s.DrawScaledText(screen, priceTag, x+s.squareSize/2-14, y+s.squareSize/2-10, s.squareSize*0.22, color.RGBA{210, 160, 50, 255})
			}

			if piece, occupied := (*bb.Board)[absoluteLoc]; occupied {
				if sprite := GetPreScaledSprite(piece); sprite != nil {
					s.dragDrawOp.GeoM.Reset()
					s.dragDrawOp.ColorScale.Reset()

					s.dragDrawOp.GeoM.Translate(x, y)
					screen.DrawImage(sprite, s.dragDrawOp)
				}
			}

			if screenCol == 0 {
				displayRank := 8 - screenRow
				if playerColor == game.Black {
					displayRank = screenRow + 1
				}
				s.DrawScaledText(screen, fmt.Sprintf("%d", displayRank), s.boardX-20, y+s.squareSize/2-6, s.squareSize*0.25, color.RGBA{150, 150, 150, 255})
			}
			if screenRow == 7 {
				s.DrawScaledText(screen, fmt.Sprintf("%c", rune('a'+screenCol)), x+s.squareSize/2-4, s.boardY+s.boardSize+5, s.squareSize*0.25, color.RGBA{150, 150, 150, 255})
			}
		}
	}

	// ==========================================
	// 				Bench Renderer
	// ==========================================

	benchY := s.boardY + s.boardSize + 10
	ebitenutil.DrawRect(screen, 0, benchY, winW, benchTrayH, bColor(40, 40, 45))

	benchTextH := benchTrayH * 0.12
	s.DrawScaledText(screen, fmt.Sprintf("BENCH INVENTORY (%d items):", len(bb.Bench)), 20, benchY+5, benchTextH, color.White)

	slotPadding := 8.0
	slotSize := benchTrayH - 45
	startX := 20.0
	currentY := benchY + 30

	for idx, piece := range bb.Bench {
		slotX := startX + float64(idx)*(slotSize+slotPadding)
		if slotX+slotSize > winW-20 {
			slotX = startX + float64(idx%6)*(slotSize+slotPadding)
			currentY = benchY + 30 + slotSize + slotPadding
		}

		isThisBenchSelected := s.isBenchSelected && s.selectedBenchIndex == idx
		if isThisBenchSelected {
			borderGlow := 3.0
			ebitenutil.DrawRect(screen, slotX-borderGlow, currentY-borderGlow, slotSize+(borderGlow*2), slotSize+(borderGlow*2), color.RGBA{0, 255, 230, 255})
		}

		if sprite := GetPreScaledSprite(piece); sprite != nil {
			s.dragDrawOp.GeoM.Reset()
			s.dragDrawOp.ColorScale.Reset()

			scaleFactor := slotSize / s.squareSize
			targetSpriteSize := slotSize * 0.75

			paddingX := (slotSize - targetSpriteSize) / 2.0
			paddingY := (slotSize - targetSpriteSize) / 2.0

			s.dragDrawOp.GeoM.Scale(scaleFactor*0.75, scaleFactor*0.75)
			s.dragDrawOp.GeoM.Translate(slotX+paddingX, currentY+paddingY)

			screen.DrawImage(sprite, s.dragDrawOp)
		}
	}

	// ==========================================
	// 			Floating Drag Renderer
	// ==========================================
	footerTextH := winH * 0.022
	footerY := winH - footerTextH - 10
	ebitenutil.DrawRect(screen, 0, footerY-5, winW, winH-footerY+5, bColor(25, 25, 30))
	s.DrawScaledText(screen, fmt.Sprintf("System Log: %s", s.statusMessage), 20, footerY, footerTextH, color.RGBA{180, 180, 180, 255})

	if s.isDragging {

		mx, my := ebiten.CursorPosition()

		var floatingPiece *game.Piece
		if s.dragSource == DragFromBench && s.draggedBenchIdx < len(bb.Bench) {
			floatingPiece = &bb.Bench[s.draggedBenchIdx]
		} else if s.dragSource == DragFromBoard {
			if piece, exists := (*bb.Board)[s.draggedBoardSquare]; exists {
				floatingPiece = &piece
			}
		}

		if floatingPiece != nil {
			if sprite := GetPreScaledSprite(*floatingPiece); sprite != nil {
				s.dragDrawOp.GeoM.Reset()
				s.dragDrawOp.ColorScale.Reset()

				// Draw the pre-scaled texture skin instantly matching cursor coordinates
				halfDim := s.squareSize / 2.0
				s.dragDrawOp.GeoM.Translate(float64(mx)-halfDim, float64(my)-halfDim)
				s.dragDrawOp.ColorScale.Scale(1, 1, 1, 0.80)

				screen.DrawImage(sprite, s.dragDrawOp)
			}
		}
	}
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
	if outsideWidth != s.lastLayoutW || outsideHeight != s.lastLayoutH {
		s.lastLayoutW = outsideWidth
		s.lastLayoutH = outsideHeight

		winW := float64(outsideWidth)
		winH := float64(outsideHeight)
		shopTrayH := winH * 0.18
		benchTrayH := winH * 0.18
		usableBoardH := winH - shopTrayH - benchTrayH - 60

		sqSize := usableBoardH / 8
		if (winW-100)/8 < sqSize {
			sqSize = (winW - 100) / 8
		}

		RegenerateScaledUICache(sqSize)
	}
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
