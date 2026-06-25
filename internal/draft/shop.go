package draft

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/CoupDeGrace92/candidates/internal/game"
)

type ItemType string

const (
	ItemPiece         ItemType = "piece"
	ItemSquare        ItemType = "square"
	ItemEngineUpgrade ItemType = "engine_upgrade"
)

type ShopItem struct {
	ID       string   `json:"id"`
	Type     ItemType `json:"type"`
	Cost     int      `json:"cost"`
	IsBought bool     `json:"is_bought"`

	PieceType    game.PieceType `json:"piece_type,omitempty"`
	UnlockSquare game.Location  `json:"unlock_square,omitempty"`
}

type ShopTray struct {
	Units   []ShopItem `json:"units"`
	Squares []ShopItem `json:"squares"`
}

type UnitBag map[game.PieceType]int

type DraftManager struct {
	rng *rand.Rand
	Bag UnitBag
}

func NewDraftManager(playerCount int) *DraftManager {
	if playerCount < 1 {
		playerCount = 1
	}

	startingBag := UnitBag{
		game.Pawn:   16 * playerCount,
		game.Knight: 6 * playerCount,
		game.Bishop: 6 * playerCount,
		game.Rook:   4 * playerCount,
		game.Queen:  2 * playerCount,
	}

	return &DraftManager{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		Bag: startingBag,
	}
}

const (
	CostPawn   = 1
	CostKnight = 3
	CostBishop = 3
	CostRook   = 5
	CostQueen  = 9

	CostRank3 = 1
	CostRank4 = 2
	CostRank5 = 5
	CostRank6 = 10

	CostReroll = 1

	MaxPieceSlots  = 4
	MaxSquareSlots = 3
)

func (dm *DraftManager) RecycleTray(tray *ShopTray) {
	if tray == nil {
		return
	}

	for _, item := range tray.Units {
		if !item.IsBought && item.Type == ItemPiece {
			dm.Bag[item.PieceType]++
		}
	}
}

func (dm *DraftManager) GenerateFreshTray(allowedSquares map[game.Location]struct{}) ShopTray {
	tray := ShopTray{
		Units:   make([]ShopItem, 3),
		Squares: make([]ShopItem, 2),
	}

	for i := 0; i < 3; i++ {
		pType := dm.drawRandomPieceFromGlobalBag()

		var cost int
		switch pType {
		case game.Pawn:
			cost = CostPawn
		case game.Knight:
			cost = CostKnight
		case game.Bishop:
			cost = CostBishop
		case game.Rook:
			cost = CostRook
		case game.Queen:
			cost = CostQueen
		}

		tray.Units[i] = ShopItem{
			ID:        fmt.Sprintf("unit_slot_%d_%d", i, dm.rng.Int63()),
			Type:      ItemPiece,
			Cost:      cost,
			PieceType: pType,
		}
	}

	//Square selection here - we want it to be connected to the players current board, not just random squares.  Rank3 more likely than rank 5 etc.
	eligibleSquares := GetEligibleSquares(allowedSquares)
	for i := 0; i < 2; i++ {
		if len(eligibleSquares) == 0 {
			break
		}

		totalWeight := 0
		for _, loc := range eligibleSquares {
			switch loc.Rank {
			case 3:
				totalWeight += 50
			case 4:
				totalWeight += 30
			case 5:
				totalWeight += 15
			case 6:
				totalWeight += 5
			default:
				totalWeight += 1
			}
		}

		roll := dm.rng.Intn(totalWeight)
		currentWeightSum := 0
		chosenIndex := 0

		for idx, loc := range eligibleSquares {
			weight := 1
			switch loc.Rank {
			case 3:
				weight = 50
			case 4:
				weight = 30
			case 5:
				weight = 15
			case 6:
				weight = 5
			}

			currentWeightSum += weight
			if roll < currentWeightSum {
				chosenIndex = idx
				break
			}
		}

		chosenLoc := eligibleSquares[chosenIndex]
		cost := GetSquareCost(chosenLoc.Rank)

		tray.Squares[i] = ShopItem{
			ID:           fmt.Sprintf("square_slot_%d_%d", i, dm.rng.Int63()),
			Type:         ItemSquare,
			Cost:         cost,
			UnlockSquare: chosenLoc,
		}

		eligibleSquares = append(eligibleSquares[:chosenIndex], eligibleSquares[chosenIndex+1:]...)
	}
	return tray
}

func GetSquareCost(rank int) int {
	switch rank {
	case 3:
		return CostRank3
	case 4:
		return CostRank4
	case 5:
		return CostRank5
	case 6:
		return CostRank6
	default:
		return 999 // Fallback safety catch for unmappable ranks
	}
}

func (dm *DraftManager) drawRandomPieceFromGlobalBag() game.PieceType {
	var pool []game.PieceType
	for pType, count := range dm.Bag {
		for i := 0; i < count; i++ {
			pool = append(pool, pType)
		}
	}

	if len(pool) == 0 {
		return game.Pawn
	}

	chosenIndex := dm.rng.Intn(len(pool))
	chosenPiece := pool[chosenIndex]

	dm.Bag[chosenPiece]--
	return chosenPiece
}

func (dm *DraftManager) ProcessReroll(tray *ShopTray, profile *game.PlayerProfile) error {
	if profile == nil || tray == nil {
		return fmt.Errorf("nil pointers passed to reroll interface boundary")
	}
	if profile.Gold < CostReroll {
		return fmt.Errorf("insufficient gold balance: need %d, have %d", CostReroll, profile.Gold)
	}

	dm.RecycleTray(tray)
	profile.Gold -= CostReroll

	*tray = dm.GenerateFreshTray(profile.BoardAndBench.Squares)
	return nil
}

func GetEligibleSquares(allowedSquares map[game.Location]struct{}) []game.Location {
	eligibleMap := make(map[game.Location]struct{})

	dx := []int{-1, 0, 1, -1, 1, -1, 0, 1}
	dy := []int{-1, -1, -1, 0, 0, 1, 1, 1}

	// 1. All squares on Rank 3 are automatically available for selection at start
	for file := 1; file <= 8; file++ {
		loc := game.Location{File: file, Rank: 3}
		if _, owned := allowedSquares[loc]; !owned {
			eligibleMap[loc] = struct{}{}
		}
	}

	for loc := range allowedSquares {
		for i := 0; i < 8; i++ {
			neighbor := game.Location{
				File: loc.File + dx[i],
				Rank: loc.Rank + dy[i],
			}

			if neighbor.File >= 1 && neighbor.File <= 8 && neighbor.Rank >= 4 && neighbor.Rank <= 6 {
				if _, owned := allowedSquares[neighbor]; !owned {
					eligibleMap[neighbor] = struct{}{}
				}
			}
		}
	}

	var list []game.Location
	for loc := range eligibleMap {
		list = append(list, loc)
	}
	return list
}

func (dm *DraftManager) BuyItem(tray *ShopTray, itemID string, profile *game.PlayerProfile) error {
	if tray == nil || profile == nil || profile.BoardAndBench == nil {
		return errors.New("nil parameters passed to transaction boundary")
	}

	var targetItem *ShopItem
	var itemListRef *[]ShopItem
	var itemIndex int

	for i, c := range tray.Units {
		if c.ID == itemID && !c.IsBought {
			targetItem = &tray.Units[i]
			itemListRef = &tray.Units
			itemIndex = i
			break
		}
	}

	if targetItem == nil {
		for i, c := range tray.Squares {
			if c.ID == itemID && !c.IsBought {
				targetItem = &tray.Squares[i]
				itemListRef = &tray.Squares
				itemIndex = i
			}
		}
	}

	if targetItem == nil {
		return fmt.Errorf("card %s is not available or has already been purchased", itemID)
	}

	if profile.Gold < targetItem.Cost {
		return fmt.Errorf("insufficient gold balance: need %d, have %d", targetItem.Cost, profile.Gold)
	}

	profile.Gold -= targetItem.Cost
	targetItem.IsBought = true
	bb := profile.BoardAndBench

	switch targetItem.Type {
	case ItemPiece:
		newPiece := game.Piece{
			Type:  targetItem.PieceType,
			Color: game.PieceColor(profile.PlayerID),
		}
		bb.Bench = append(bb.Bench, newPiece)
	case ItemSquare:
		bb.Squares[targetItem.UnlockSquare] = struct{}{}
	case ItemEngineUpgrade:
		if profile.SkillLevel < 20 {
			profile.SkillLevel++
		}
	}

	*itemListRef = append((*itemListRef)[:itemIndex], (*itemListRef)[itemIndex+1:]...)
	return nil
}
