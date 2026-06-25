package draft

import (
	"math/rand"
	"testing"

	"github.com/CoupDeGrace92/candidates/internal/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuyItem_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		cardID        string
		initialGold   int
		setupTray     func() ShopTray
		verifyResults func(t *testing.T, tray *ShopTray, profile *game.PlayerProfile, err error)
	}{
		{
			name:        "Successfully Purchase Piece - Moves to Bench, Deducts Gold, Removes Card",
			cardID:      "unit_1",
			initialGold: 10,
			setupTray: func() ShopTray {
				return ShopTray{
					Units: []ShopItem{
						{ID: "unit_1", Type: ItemPiece, Cost: CostKnight, PieceType: game.Knight},
						{ID: "unit_2", Type: ItemPiece, Cost: CostPawn, PieceType: game.Pawn},
					},
					Squares: []ShopItem{},
				}
			},
			verifyResults: func(t *testing.T, tray *ShopTray, profile *game.PlayerProfile, err error) {
				require.NoError(t, err)
				// 10 Gold starting - 3 Gold Minor Piece Cost = 7 Gold remaining
				assert.Equal(t, 7, profile.Gold)

				// Card must be appended to the tightly coupled bench slice
				require.Len(t, profile.BoardAndBench.Bench, 1)
				assert.Equal(t, game.Knight, profile.BoardAndBench.Bench[0].Type)

				// Slicing check: Card should be physically erased from the view tray
				assert.Len(t, tray.Units, 1)
				assert.Equal(t, "unit_2", tray.Units[0].ID)
			},
		},
		{
			name:        "Successfully Purchase Adjacent Square - Updates Map, Deducts Gold, Removes Card",
			cardID:      "square_1",
			initialGold: 5,
			setupTray: func() ShopTray {
				return ShopTray{
					Units: []ShopItem{},
					Squares: []ShopItem{
						{ID: "square_1", Type: ItemSquare, Cost: 2, UnlockSquare: game.Location{File: 4, Rank: 3}}, // d3
					},
				}
			},
			verifyResults: func(t *testing.T, tray *ShopTray, profile *game.PlayerProfile, err error) {
				require.NoError(t, err)
				assert.Equal(t, 3, profile.Gold)

				// Verify the dynamic coordinate snaps directly into the player's unique Set
				targetLoc := game.Location{File: 4, Rank: 3}
				_, owned := profile.BoardAndBench.Squares[targetLoc]
				assert.True(t, owned)

				// Card should be completely wiped from the squares slice array
				assert.Len(t, tray.Squares, 0)
			},
		},
		{
			name:        "Successfully Purchase Engine Upgrade - Increments Skill Level",
			cardID:      "upgrade_1",
			initialGold: 5,
			setupTray: func() ShopTray {
				return ShopTray{
					Units: []ShopItem{
						{ID: "upgrade_1", Type: ItemEngineUpgrade, Cost: 3},
					},
					Squares: []ShopItem{},
				}
			},
			verifyResults: func(t *testing.T, tray *ShopTray, profile *game.PlayerProfile, err error) {
				require.NoError(t, err)
				assert.Equal(t, 2, profile.Gold)
				// Starting default config is 7, upgrade bumps it safely to 8
				assert.Equal(t, 8, profile.SkillLevel)
				assert.Len(t, tray.Units, 0)
			},
		},
		{
			name:        "Fail Purchase - Insufficient Gold Balance",
			cardID:      "expensive_queen",
			initialGold: 2, // Too poor to afford a Queen
			setupTray: func() ShopTray {
				return ShopTray{
					Units: []ShopItem{
						{ID: "expensive_queen", Type: ItemPiece, Cost: CostQueen, PieceType: game.Queen},
					},
					Squares: []ShopItem{},
				}
			},
			verifyResults: func(t *testing.T, tray *ShopTray, profile *game.PlayerProfile, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "insufficient gold balance")
				assert.Equal(t, 2, profile.Gold) // Currency remains untouched
				assert.Len(t, profile.BoardAndBench.Bench, 0)
				assert.Len(t, tray.Units, 1) // Card remains on offer
			},
		},
		{
			name:        "Fail Purchase - Card Missing from Active Tray",
			cardID:      "non_existent_id",
			initialGold: 10,
			setupTray: func() ShopTray {
				return ShopTray{
					Units:   []ShopItem{{ID: "unit_1", Type: ItemPiece, Cost: CostPawn, PieceType: game.Pawn}},
					Squares: []ShopItem{},
				}
			},
			verifyResults: func(t *testing.T, tray *ShopTray, profile *game.PlayerProfile, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "is not available or has already been purchased")
				assert.Equal(t, 10, profile.Gold)
				assert.Len(t, tray.Units, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			dm := &DraftManager{rng: rand.New(rand.NewSource(1))}
			profile := game.NewDefaultProfile("white", true)
			profile.Gold = tt.initialGold

			if profile.BoardAndBench.Squares == nil {
				profile.BoardAndBench.Squares = make(map[game.Location]struct{})
			}

			tray := tt.setupTray()

			// Act
			err := dm.BuyItem(&tray, tt.cardID, profile)

			// Assert
			tt.verifyResults(t, &tray, profile, err)
		})
	}
}

func TestGetEligibleSquares_Filtering(t *testing.T) {
	// Arrange - Baseline player state natively owns Ranks 1 and 2
	profile := game.NewDefaultProfile("white", true)

	// Act
	eligible := GetEligibleSquares(profile.BoardAndBench.Squares)

	// Assert
	// At game start, exactly all 8 squares of Rank 3 should be eligible
	assert.Len(t, eligible, 8)
	for _, loc := range eligible {
		assert.Equal(t, 3, loc.Rank)
	}

	// Manually unlock a square on Rank 3 to test expansion upward into Rank 4
	profile.BoardAndBench.Squares[game.Location{File: 4, Rank: 3}] = struct{}{} // d3 owned

	// Re-evaluate adjacency paths
	newEligible := GetEligibleSquares(profile.BoardAndBench.Squares)

	// The map should now contain unowned Rank 3 squares AND neighbor expansions onto Rank 4 (c4, d4, e4)
	hasD4 := false
	for _, loc := range newEligible {
		if loc.File == 4 && loc.Rank == 4 {
			hasD4 = true
		}
		// Confirm bounds constraints restrict options below Rank 7
		assert.True(t, loc.Rank <= 6)
	}
	assert.True(t, hasD4, "Expansion path failed to unlock neighboring Rank 4 tile d4")
}
