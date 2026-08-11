package scene

import (
	"context"
	"fmt"
	"image/color"
	"log"

	"github.com/CoupDeGrace92/candidates/internal/engine"
	"github.com/CoupDeGrace92/candidates/internal/game"
	"github.com/CoupDeGrace92/candidates/internal/tournament"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type BattleScene struct {
	matchState *game.MatchState
	history    *game.PositionTracker
	drawish    int
	controller *engine.MatchController
	simCtx     context.Context
	cancelSim  context.CancelFunc
	moveChan   <-chan engine.MoveResult

	// Animation and Flow Control
	isCalculating    bool
	accumulatedTicks int
	ticksPerMove     int
	statusMessage    string

	// --- TOURNAMENT LIFECYCLE ROUTER CONNECTER ---
	director       *Director // Added to handle state routing post-battle
	matchConcluded bool      // Flag to protect against duplicate submission calls
	// ----------------------------------------------
}

// NewBattleScene instantiates your production battle scene simulation canvas.
// Updated parameter inputs to accept your Director pointer instance.
func NewBattleScene(binPath string, initialState *game.MatchState, director *Director) (*BattleScene, error) {
	pCfg := engine.Config{
		SkillLevel: initialState.WhitePlayer.SkillLevel,
		MoveTimeMs: initialState.WhitePlayer.MoveTimeMs,
	}
	eCfg := engine.Config{
		SkillLevel: initialState.BlackPlayer.SkillLevel,
		MoveTimeMs: initialState.BlackPlayer.MoveTimeMs,
	}

	ctrl, err := engine.NewMatchController(binPath, pCfg, eCfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	hist := make(game.PositionTracker)

	// Automatically calculate and inject castling rights flags right before process initialization
	initialState.InitializeCastlingRights()

	return &BattleScene{
		matchState:     initialState,
		history:        &hist,
		controller:     ctrl,
		simCtx:         ctx,
		cancelSim:      cancel,
		ticksPerMove:   25,
		statusMessage:  "Simulation starting...",
		director:       director,
		matchConcluded: false,
	}, nil
}

func (b *BattleScene) Update() error {
	// ==================================================
	// INTERCEPT GAME OVER KEY-TAP RETURN GATES
	// ==================================================
	// Once Stockfish determines a result, pressing the SPACEBAR wraps the
	// game's moves into a PGN data frame and passes it back to the tournament manager.
	if b.matchConcluded {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			log.Printf("[Battle] Compiling final match results frame for the tournament.")

			// Build a clean outcome payload
			winner := ""
			isDraw := true
			if b.statusMessage == "Game Over: White achieves victory!" {
				winner = b.matchState.WhitePlayer.PlayerID
				isDraw = false
			} else if b.statusMessage == "Game Over: Black achieves victory!" {
				winner = b.matchState.BlackPlayer.PlayerID
				isDraw = false
			}

			var activePairingID string
			if pairings, err := b.director.tourney.GeneratePairings(); err == nil && len(pairings) > 0 {
				activePairingID = pairings[0].ID
			} else {
				// Pure procedural fallback string token if the array check fails
				activePairingID = fmt.Sprintf("p_r%d_%s_vs_%s",
					b.director.tourney.CurrentRound(),
					b.matchState.WhitePlayer.PlayerID,
					b.matchState.BlackPlayer.PlayerID,
				)
			}

			outcome := tournament.MatchOutcome{
				PairingID: activePairingID,
				WinnerID:  winner,
				IsDraw:    isDraw,
				PGNData:   fmt.Sprintf("[White %s] [Black %s] [Result %s]", b.matchState.WhitePlayer.PlayerID, b.matchState.BlackPlayer.PlayerID, b.statusMessage),
			}

			// Cleanly terminate active engine pipes before leaving the room
			b.Destroy()

			// Return to the shop
			b.director.ConcludeBattleAndReturnToShop(outcome)
			return nil
		}
		return nil
	}

	if b.isCalculating {
		select {
		case result, open := <-b.moveChan:
			if !open {
				b.isCalculating = false
				return nil
			}

			if result.Status == engine.StatusCheckmate {
				// Identify who delivered the checkmate depending on who just moved
				if b.matchState.ActiveColor == game.White {
					b.statusMessage = "Game Over: White achieves victory!"
				} else {
					b.statusMessage = "Game Over: Black achieves victory!"
				}
				b.isCalculating = false
				b.matchConcluded = true
				return nil
			}
			if result.Status == engine.StatusStalemate {
				b.statusMessage = "Game Over: Draw by Stalemate!"
				b.isCalculating = false
				b.matchConcluded = true
				return nil
			}

			if result.IsEngineDraw {
				b.drawish++
				if b.drawish > 10 {
					b.statusMessage = "Game Over: Draw by Agreement!"
					b.isCalculating = false
					b.matchConcluded = true
					return nil
				}
			} else {
				b.drawish = 0
			}

			if result.Err != nil {
				b.statusMessage = fmt.Sprintf("Engine error: %v", result.Err)
				b.isCalculating = false
				b.matchConcluded = true
				return nil
			}

			_ = b.matchState.ApplyMove(result.Move)
			if b.history.RecordPosition(b.matchState.ToFEN()) {
				b.statusMessage = "Game Over: Draw by Threefold Repetition!"
				b.isCalculating = false
				b.matchConcluded = true
				return nil
			}

			if b.matchState.HalfMoveClock >= 100 {
				b.statusMessage = "Game Over: Draw by 50-move rule limit"
				b.isCalculating = false
				b.matchConcluded = true
				return nil
			}
		default:
			// Stockfish is still thinking
		}
		return nil
	}

	b.accumulatedTicks++
	if b.accumulatedTicks >= b.ticksPerMove {
		b.isCalculating = true
		b.accumulatedTicks = 0
		b.moveChan = b.controller.SimNextTurn(b.simCtx, b.matchState)
	}

	return nil
}

func (b *BattleScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 30, 255})

	squareSize := 60
	offsetX := 40
	offsetY := 60

	for rank := 8; rank >= 1; rank-- {
		for file := 1; file <= 8; file++ {
			x := offsetX + (file-1)*squareSize
			y := offsetY + (8-rank)*squareSize

			var activeTheme game.BoardTheme
			if rank <= 4 {
				activeTheme = b.matchState.WhitePlayer.Theme
			} else {
				activeTheme = b.matchState.BlackPlayer.Theme
			}

			rectColor := activeTheme.LightSquare
			if (rank+file)%2 == 0 {
				rectColor = activeTheme.DarkSquare
			}

			ebitenutil.DrawRect(screen, float64(x), float64(y), float64(squareSize-2), float64(squareSize-2), rectColor)

			loc := game.Location{File: file, Rank: rank}

			// --------------------------------------------------
			// FIXED BOARD DRAWING: NO REAL-TIME CONCATENATION
			// --------------------------------------------------
			// We query your unified MatchState.Board map directly, dropping
			// the broken cross-profile lookups completely.
			if piece, occupied := b.matchState.Board[loc]; occupied {
				if sprite := GetPieceSprite(piece); sprite != nil {
					op := &ebiten.DrawImageOptions{}

					scaleX := float64(squareSize) / float64(SpriteW)
					scaleY := float64(squareSize) / float64(SpriteH)

					op.GeoM.Scale(scaleX, scaleY)
					op.GeoM.Translate(float64(x), float64(y))

					screen.DrawImage(sprite, op)
				}
			}
		}
	}

	ebitenutil.DebugPrintAt(screen, "THE CANDIDATES - RESOLUTION PHASE", 40, 20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Status: %s", b.statusMessage), 40, 560)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Turn Num: %d (%s to move)", b.matchState.FullMoveNumber, b.matchState.ActiveColor), 40, 580)

	// Append an intuitive prompt when the game has completed
	if b.matchConcluded {
		ebitenutil.DebugPrintAt(screen, ">> PRESS SPACEBAR TO CLAIM REWARDS AND RETURN TO SHOP <<", 40, 600)
	}
}

func (b *BattleScene) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 640, 640
}

func (b *BattleScene) Destroy() {
	b.cancelSim()
	b.controller.Terminate()
	log.Println("Battlescene terminated, background engines cleanly harvested")
}
