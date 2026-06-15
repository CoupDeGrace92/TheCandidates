package engine

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Config struct {
	SkillLevel int //This value is the targetted ELO of the engine
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

	return instance, nil
}

func (s *StockfishInstance) UpdateConfig(cfg Config) {
	s.config = &cfg
	s.sendCommand(fmt.Sprintf("setoption name Skill Level value %d", cfg.SkillLevel))
}

func (s *StockfishInstance) RequestMove(fen string) (string, error) {
	s.sendCommand(fmt.Sprintf("position fen %s", fen))
	s.sendCommand(fmt.Sprintf("go movetime %d", s.config.MoveTimeMs))
	scanner := bufio.NewScanner(s.stdout)

	//Listen to the text output stream synchronously until we find "bestmove"
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				return parts[1], nil
			}
			return "", fmt.Errorf("malformed bestmove output from engine: %s", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading engine stdout stream: %w", err)
	}

	return "", fmt.Errorf("engine closed output prematurely")
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
