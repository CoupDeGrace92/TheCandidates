package scene

import (
	"fmt"
	"image/color"
	"log"

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

	//Temp reroll button with collision boundaries
	rerollBtn imageRect
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
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if s.rerollBtn.Contains(mx, my) {
			if err := s.manager.ProcessReroll(&s.tray, s.profile); err != nil {
				s.statusMessage = fmt.Sprintf("Reroll Failed: %v", err)
			} else {
				s.statusMessage = "Shop refreshed! Piece bag has been adjusted"
			}
			return nil
		}

		for _, item := range s.tray.Units {
			itemRect := s.getUnitItemRect(item.ID)
			if itemRect.Contains(mx, my) {
				if err := s.manager.BuyItem(&s.tray, item.ID, s.profile); err != nil {
					s.statusMessage = fmt.Sprintf("Purchase unsuccesful: %v", err)
				} else {
					s.statusMessage = "Unit succesfully drafted to your bench"
				}
				return nil
			}
		}

		for _, item := range s.tray.Squares {
			itemRect := s.getSquareItemRect(item.ID)
			if itemRect.Contains(mx, my) {
				if err := s.manager.BuyItem(&s.tray, item.ID, s.profile); err != nil {
					s.statusMessage = fmt.Sprintf("Purchase unsuccesful: %v", err)
				} else {
					s.statusMessage = "Square aquired!"
				}
				return nil
			}
		}
	}
	return nil
}

func (s *ShopScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{25, 25, 30, 255})

	ebitenutil.DebugPrintAt(screen, "THE CANDIDATES - DRAFTING SHOP", 40, 20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("YOUR GOLD: %d G", s.profile.Gold), 40, 50)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ENGINE LEVEL: %d / 20", s.profile.SkillLevel), 240, 50)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("BENCH UNITS: %d", len(s.profile.BoardAndBench.Bench)), 440, 50)

	ebitenutil.DebugPrintAt(screen, "--- UNITS FOR SALE ---", 40, 100)
	for _, card := range s.tray.Units {
		r := s.getUnitItemRect(card.ID)
		ebitenutil.DrawRect(screen, r.x, r.y, r.w, r.h, color.RGBA{45, 45, 55, 255})

		nameStr := fmt.Sprintf("%s\nCost: %dG", card.PieceType, card.Cost)
		ebitenutil.DebugPrintAt(screen, nameStr, int(r.x+10), int(r.y+20))
	}

	ebitenutil.DebugPrintAt(screen, "--- TERRITORY EXPANSIONS ---", 360, 100)
	for _, card := range s.tray.Squares {
		r := s.getSquareItemRect(card.ID)
		ebitenutil.DrawRect(screen, r.x, r.y, r.w, r.h, color.RGBA{55, 45, 45, 255})

		fileChar := rune('a' + card.UnlockSquare.File - 1)
		nameStr := fmt.Sprintf("Square: %c%d\nCost: %dG", fileChar, card.UnlockSquare.Rank, card.Cost)
		ebitenutil.DebugPrintAt(screen, nameStr, int(r.x+10), int(r.y+20))
	}

	ebitenutil.DrawRect(screen, s.rerollBtn.x, s.rerollBtn.y, s.rerollBtn.w, s.rerollBtn.h, color.RGBA{140, 110, 40, 255})
	ebitenutil.DebugPrintAt(screen, "REROLL (1G)", int(s.rerollBtn.x+20), int(s.rerollBtn.y+12))

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Log: %s", s.statusMessage), 40, 570)
}

func (s *ShopScene) Layout(outsideWidth, outsideHeigh int) (int, int) {
	return 640, 640
}

func (s *ShopScene) getUnitItemRect(itemID string) imageRect {
	for i, c := range s.tray.Units {
		if c.ID == itemID {
			return imageRect{x: float64(40 + i*90), y: 130, w: 85, h: 85}
		}
	}
	return imageRect{}
}

func (s *ShopScene) getSquareItemRect(itemID string) imageRect {
	for i, c := range s.tray.Squares {
		if c.ID == itemID {
			return imageRect{x: float64(360 + i*115), y: 130, w: 100, h: 100}
		}
	}
	return imageRect{}
}

func (s *ShopScene) Destroy() {
	// For now, this acts as a placeholder for saving state or releasing textures.
	log.Println("Shop Scene destroyed, draft variables safely finalized.")
}
