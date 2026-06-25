package engine

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Config struct {
	SkillLevel int //This value is the targetted ELO of the engine, currently configured for stockfishes SkillLevel, can be reconfiged for the ELO command
	MoveTimeMs int
}

type StockfishInstance struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	config *Config
}

func NewInstance(binaryPath string, cfg Config) (*StockfishInstance, error) {
	cmd := exec.Command(binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start stockfish executable")
	}

	instance := &StockfishInstance{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		config: &cfg,
	}

	instance.sendCommand("uci")
	instance.UpdateConfig(cfg)
	instance.sendCommand("isready")

	hasReadyOK := false

	//We must flush the stdout pipe before we start feeding moves:
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "info string CRITICAL ERROR") {
			return nil, fmt.Errorf("stockfish booted with fatal interanl configuration error: %s", line)
		}

		if line == "readyok" {
			hasReadyOK = true
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stockfish initialization handshake")
	}

	if !hasReadyOK {
		return nil, fmt.Errorf("stockfish process terminated or closed prematurely")
	}

	return instance, nil
}

func (s *StockfishInstance) UpdateConfig(cfg Config) {
	s.config = &cfg
	s.sendCommand(fmt.Sprintf("setoption name Skill Level value %d", cfg.SkillLevel))
}

func (s *StockfishInstance) RequestMove(fen string) (MoveResult, error) {
	s.sendCommand(fmt.Sprintf("position fen %s", fen))
	s.sendCommand(fmt.Sprintf("go movetime %d", s.config.MoveTimeMs))
	scanner := bufio.NewScanner(s.stdout)

	//Listen to the text output stream synchronously until we find "bestmove"
	var result MoveResult

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "info") && strings.Contains(line, "score") {
			parts := strings.Split(line, " ")
			for i, part := range parts {
				if part == "score" && i+2 < len(parts) {
					scoreType := parts[i+1]
					scoreValStr := parts[i+2]

					var scoreVal int
					_, _ = fmt.Sscanf(scoreValStr, "%d", &scoreVal)

					if scoreType == "mate" {
						result.ScoreMateIn = scoreVal
					} else if scoreType == "cp" && scoreVal == 0 {
						result.IsEngineDraw = true
					} else {
						result.IsEngineDraw = false
					}
				}
			}
		}

		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				result.Move = parts[1]

				if result.Move == "(none)" || result.Move == "0000" {
					if result.ScoreMateIn != 0 {
						result.Status = StatusStalemate
					} else {
						result.Status = StatusCheckmate
					}
				}
			}
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return MoveResult{}, fmt.Errorf("error reading engine stdout stream: %w", err)
	}

	if result.Move == "" {
		return MoveResult{}, fmt.Errorf("Engine stream ended or halted without delivering a move")
	}
	return result, nil
}

func (s *StockfishInstance) Close() {
	s.sendCommand("quit")
	_ = s.stdin.Close()
	_ = s.stdout.Close()
	_ = s.cmd.Wait()
}

// We might want to look here to see if the send command can error in non-graceful ways
func (s *StockfishInstance) sendCommand(cmd string) {
	_, _ = fmt.Fprintln(s.stdin, cmd)
}
