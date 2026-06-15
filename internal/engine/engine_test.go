package engine

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/CoupDeGrace92/candidates/internal/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestBinaryPath(t *testing.T) string {
	path := "../../assets/engines/stockfish"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Stockfish testing binary not found at ./testbin/stockfish")
	}
	return path
}

func TestStockfishInstance_LifecycleAndMove(t *testing.T) {
	binPath := getTestBinaryPath(t)
	cfg := Config{SkillLevel: 10, MoveTimeMs: 50}

	instance, err := NewInstance(binPath, cfg)
	require.NoError(t, err)
	require.NotNil(t, instance)

	assert.NotNil(t, instance.cmd)
	assert.NotNil(t, instance.stdin)
	assert.NotNil(t, instance.stdout)

	startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	move, err := instance.RequestMove(startFEN)

	require.NoError(t, err)
	assert.NotEmpty(t, move)
	assert.Len(t, move, 4)

	newCfg := Config{SkillLevel: 15, MoveTimeMs: 100}
	assert.NotPanics(t, func() {
		instance.UpdateConfig(newCfg)
	})
	assert.Equal(t, 15, instance.config.SkillLevel)

	assert.NotPanics(t, func() {
		instance.Close()
	})
}

func TestMatchController_AsynDualSimulation(t *testing.T) {
	binPath := getTestBinaryPath(t)

	whiteCfg := Config{SkillLevel: 18, MoveTimeMs: 40}
	blackCfg := Config{SkillLevel: 5, MoveTimeMs: 40}

	controller, err := NewMatchController(binPath, whiteCfg, blackCfg)
	require.NoError(t, err)
	require.NotNil(t, controller)
	defer controller.Terminate()

	wbs := make(game.BoardState)
	bbs := make(game.BoardState)

	wbs[game.Location{File: 5, Rank: 2}] = game.Piece{Type: game.Pawn, Color: game.White}
	wbs[game.Location{File: 5, Rank: 1}] = game.Piece{Type: game.King, Color: game.White}

	bbs[game.Location{File: 5, Rank: 8}] = game.Piece{Type: game.King, Color: game.Black}

	matchState := &game.MatchState{
		ActiveColor:     game.White,
		CastlingRights:  "-",
		EnPassantTarget: "-",
		HalfMoveClock:   0,
		FullMoveNumber:  1,
		WhitePlayer:     &game.PlayerProfile{Board: &wbs},
		BlackPlayer:     &game.PlayerProfile{Board: &bbs},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resultChan := controller.SimNextTurn(ctx, matchState)
	require.NotNil(t, resultChan)

	var result MoveResult
	select {
	case res, open := <-resultChan:
		require.True(t, open, "Channel closed prematurely without sending move")
		result = res
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("Simulation turn calculation locked up or timed out past 1500 ms")
	}

	assert.NoError(t, result.Err)
	assert.NotEmpty(t, result.Move)
	t.Logf("Asynchronous Engine returned move :%s", result.Move)
}
