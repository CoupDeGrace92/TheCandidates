package scene

import (
	"context"
	"fmt"
	"image/color"
	"log"

	"github.com/CoupDeGrace92/candidates/internal/engine"
	"github.com/CoupDeGrace92/candidates/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type BattleScene struct {
	matchState *game.MatchState
	controller *engine.MatchController
	simCtx     context.Context
	cancelSim  context.CancelFunc
	moveChan   <-chan engine.MoveResult

	//Animation and FLow Control
	isCalculating    bool
	accumulatedTicks int
	ticksPerMove     int
	statusMessage    string
}

func NewBattleScene(binPath string, initialState *game.MatchState) (*BattleScene, error) {
	//These defaults are currently hardwired
	//TODO - MAKE THESE CONFIGURABLE
	playerCfg := engine.Config{SkillLevel: 18, MoveTimeMs: 150}
	enemyCfg := engine.Config{SkillLevel: 12, MoveTimeMs: 150}

	ctrl, err := engine.NewMatchController(binPath, playerCfg, enemyCfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &BattleScene{
		matchState:    initialState,
		controller:    ctrl,
		simCtx:        ctx,
		cancelSim:     cancel,
		ticksPerMove:  45,
		statusMessage: "Simulation starting...", //again a placeholder: TODO replace this
	}, nil

}

func (b *BattleScene) Update() error {
	if b.isCalculating {
		select {
		case result, open := <-b.moveChan:
			if !open {
				b.isCalculating = false
				return nil
			}

			if result.Err != nil {
				b.statusMessage = fmt.Sprintf("Engine error: %v", result.Err)
				b.isCalculating = false
				return nil
			}

			b.statusMessage = fmt.Sprintf("%s plays: %s", b.matchState.ActiveColor, result.Move)
			if err := b.matchState.ApplyMove(result.Move); err != nil {
				b.statusMessage = fmt.Sprintf("Move mutation error: %v", err)
				b.isCalculating = false
				return nil
			}

			b.isCalculating = false
			b.accumulatedTicks = 0
		default:
			//Stockfish is still thinking
		}
		return nil
	}

	//Waiting x ticks to make the game watchable before calling stockfish again
	b.accumulatedTicks++
	if b.accumulatedTicks >= b.ticksPerMove {
		b.isCalculating = true
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

			rectColor := color.RGBA{235, 235, 235, 255}
			if (rank+file)%2 == 0 {
				rectColor = color.RGBA{120, 135, 120, 255}
			}

			ebitenutil.DrawRect(screen, float64(x), float64(y), float64(squareSize-2), float64(squareSize-2), rectColor)

			loc := game.Location{File: file, Rank: rank}
			combined := game.ConcatenateBoardState(b.matchState.WhitePlayer.Board, b.matchState.BlackPlayer.Board)
			if piece, occupied := (*combined)[loc]; occupied {
				char := "?"
				switch piece.Type {
				case game.Pawn:
					char = "P"
				case game.Knight:
					char = "N"
				case game.Bishop:
					char = "B"
				case game.Rook:
					char = "R"
				case game.Queen:
					char = "Q"
				case game.King:
					char = "K"
				}

				colorTag := "w"
				if piece.Color == game.Black {
					colorTag = "b"
				}
				displayStr := fmt.Sprintf("%s.%s", char, colorTag)

				ebitenutil.DebugPrintAt(screen, displayStr, x+15, y+22)

			}
		}
	}

	ebitenutil.DebugPrintAt(screen, "THE CANDIDATES - RESOLUTION PHASE", 40, 20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Status: %s", b.statusMessage), 40, 560)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Turn Num: %d", b.matchState.FullMoveNumber), 40, 580)
}

// setting resolution independant of screen sizing boundaries
func (b *BattleScene) Layout(outsideWidth, outsideHeigh int) (int, int) {
	return 640, 640
}

func (b *BattleScene) Destroy() {
	b.cancelSim()
	b.controller.Terminate()
	log.Println("Battlescene terminated, background engines cleanly harvested")
}
