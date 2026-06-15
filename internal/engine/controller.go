package engine

import (
	"context"
	"fmt"

	"github.com/CoupDeGrace92/candidates/internal/game"
)

type MoveResult struct {
	Move string
	Err  error
}

type MatchController struct {
	WhiteEngine *StockfishInstance
	BlackEngine *StockfishInstance
}

func NewMatchController(binPath string, whiteCfg, blackCfg Config) (*MatchController, error) {
	white, err := NewInstance(binPath, whiteCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init white engine: %w", err)
	}

	black, err := NewInstance(binPath, blackCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init black engine: %w", err)
	}

	return &MatchController{
		WhiteEngine: white,
		BlackEngine: black,
	}, nil
}

func (m *MatchController) SimNextTurn(ctx context.Context, state *game.MatchState) <-chan MoveResult {
	out := make(chan MoveResult, 1)
	currentFEN := state.ToFEN()

	go func() {
		defer close(out)

		var selectedEngine *StockfishInstance

		if state.ActiveColor == "white" {
			selectedEngine = m.WhiteEngine
		} else {
			selectedEngine = m.BlackEngine
		}

		move, err := selectedEngine.RequestMove(currentFEN)

		select {
		case <-ctx.Done():
			return
		case out <- MoveResult{Move: move, Err: err}:
		}
	}()

	return out
}

func (m *MatchController) Terminate() {
	if m.WhiteEngine != nil {
		m.WhiteEngine.Close()
	}
	if m.BlackEngine != nil {
		m.BlackEngine.Close()
	}
}
