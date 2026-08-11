package game

import (
	"fmt"
	"math/rand"
	"time"
)

// MergePlayerPlacements unifies two player draft setups into a single global board.
// Returns the compiled BoardState and a success boolean. If the referee has to invoke the
// standard 2-King fallback, success returns false so the UI can flag a warning.
func MergePlayerPlacements(whiteProfile, blackProfile *PlayerProfile) (BoardState, bool, error) {
	if whiteProfile == nil || blackProfile == nil || whiteProfile.BoardAndBench == nil || blackProfile.BoardAndBench == nil {
		return nil, false, fmt.Errorf("cannot merge nil player profile containers")
	}

	globalBoard := make(BoardState)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// ==================================================
	// 1. INTEGRATE WHITE PLAYER LAYOUT DIRECTLY
	// ==================================================
	whiteBB := whiteProfile.BoardAndBench
	ensureKingIsDeployed(whiteBB)

	for loc, piece := range *whiteBB.Board {
		piece.Color = White
		globalBoard[loc] = piece // Anchored directly on global Ranks 1-2
	}

	// ==================================================
	// 2. INTEGRATE BLACK PLAYER LAYOUT DIRECTLY
	// ==================================================
	blackBB := blackProfile.BoardAndBench
	ensureKingIsDeployed(blackBB)

	// FIXED ARCHITECTURE: Because SwitchColorIfDifferent already flipped Black's
	// keys into global Ranks 7-8 inside their profile, we copy them directly with
	// ZERO late-stage transformations! This completely eliminates the double-inversion bug.
	for loc, piece := range *blackBB.Board {
		piece.Color = Black
		globalBoard[loc] = piece // Anchored directly on global Ranks 7-8
	}

	// ==================================================
	// 3. MID-BOARD COLLISION DISPUTES (COIN FLIPS)
	// ==================================================
	// In multiplayer matches where players expand into shared central territory ranks,
	// coin-flips determine who retains ownership of overlapping tile coordinates.
	occupiedSquares := make(map[Location]struct{})
	for loc := range globalBoard {
		occupiedSquares[loc] = struct{}{}
	}

	// Re-verify the board entries from Black's profile to handle collision overrides
	for blackLoc, blackPiece := range *blackBB.Board {
		if whitePiece, collision := globalBoard[blackLoc]; collision && whitePiece.Color == White {
			if rng.Float64() < 0.5 {
				// White wins the coin flip: bump Black to an open spot in Black's territory
				// findOpenSquareForBlack handles looking up space inside Black's global parameters
				bumpLoc, placed := findOpenSquareForBlack(blackBB, occupiedSquares, rng)
				if placed {
					globalBoard[bumpLoc] = blackPiece
					occupiedSquares[bumpLoc] = struct{}{}
				} else {
					// If territory is 100% full, the piece is intentionally omitted from combat
					delete(globalBoard, blackLoc)
				}
			} else {
				// Black wins the coin flip: displace White's piece to an open slot in White's territory
				globalBoard[blackLoc] = blackPiece

				bumpLoc, placed := findOpenSquareForWhite(whiteBB, occupiedSquares, rng)
				if placed {
					globalBoard[bumpLoc] = whitePiece
					occupiedSquares[bumpLoc] = struct{}{}
				}
			}
		}
	}

	// ==================================================
	// 4. UNRESTRICTED DYNAMIC RESOLUTION MACHINE
	// ==================================================
	passAttempts := 0
	maxAttempts := 50

	for passAttempts < maxAttempts {
		legal, wk, bk, offending := LegalPosition(&globalBoard, White)
		if legal && wk && bk {
			return globalBoard, true, nil // Success! Perfect legal board configuration reached
		}

		// STEP A: FORCE-SPAWN MISSING KINGS IMMEDIATELY
		if !wk {
			spawnKingOnRandomOpenSquare(&globalBoard, White, whiteBB, rng)
			continue
		}
		if !bk {
			spawnKingOnRandomOpenSquare(&globalBoard, Black, blackBB, rng)
			continue
		}

		// STEP B: UNRESTRICTED SHIFTING & SOFT-OMISSIONS
		for _, targetLoc := range offending {
			piece, exists := globalBoard[targetLoc]
			if !exists {
				continue
			}

			exclusions := make(map[Location]struct{})
			for loc := range globalBoard {
				exclusions[loc] = struct{}{}
			}

			// Forbid pawns from landing on back ranks during scatter shuffles
			if piece.Type == Pawn {
				for f := 1; f <= 8; f++ {
					exclusions[Location{File: f, Rank: 1}] = struct{}{}
					exclusions[Location{File: f, Rank: 8}] = struct{}{}
				}
			}

			bbContext := whiteBB
			if piece.Color == Black {
				bbContext = blackBB
			}

			candidateLoc, placed := bbContext.RandomOpenSquare(exclusions)
			if placed {
				delete(globalBoard, targetLoc)
				globalBoard[candidateLoc] = piece

				testLegal, testWK, testBK, _ := LegalPosition(&globalBoard, White)
				if testLegal && testWK && testBK {
					break
				}
				delete(globalBoard, candidateLoc)
				globalBoard[targetLoc] = piece
			} else {
				delete(globalBoard, targetLoc) // Soft-omission if packed full
				break
			}
		}

		passAttempts++
	}

	// ==================================================
	// 5. THE STANDARD FAILSAFE BACKSTOP
	// ==================================================
	finalLegal, finalWK, finalBK, _ := LegalPosition(&globalBoard, White)
	if !finalLegal || !finalWK || !finalBK {
		fallbackBoard := BoardState{
			Location{File: 5, Rank: 1}: {Type: King, Color: White},
			Location{File: 5, Rank: 8}: {Type: King, Color: Black},
		}
		return fallbackBoard, false, nil // UI scene uses false to print a specific error message log
	}

	return globalBoard, true, nil
}

func spawnKingOnRandomOpenSquare(board *BoardState, color PieceColor, pp *PlayerPieces, rng *rand.Rand) {
	exclusions := make(map[Location]struct{})
	for loc := range *board {
		exclusions[loc] = struct{}{}
	}

	candidateLoc, placed := pp.RandomOpenSquare(exclusions)
	if !placed {
		var fallback []Location
		for r := 1; r <= 8; r++ {
			for f := 1; f <= 8; f++ {
				loc := Location{File: f, Rank: r}
				if _, occupied := (*board)[loc]; !occupied {
					fallback = append(fallback, loc)
				}
			}
		}
		if len(fallback) > 0 {
			candidateLoc = fallback[rng.Intn(len(fallback))]
			placed = true
		}
	}

	if placed {
		(*board)[candidateLoc] = Piece{Type: King, Color: color}
	}
}

func ensureKingIsDeployed(p *PlayerPieces) {
	kingPlaced := false
	for _, piece := range *p.Board {
		if piece.Type == King {
			kingPlaced = true
			break
		}
	}

	if !kingPlaced {
		kingIdx := -1 //While there has to be a king on the bench or board, this exists to double check
		for i, piece := range p.Bench {
			if piece.Type == King {
				kingIdx = i
				break
			}
		}

		occupied := make(map[Location]struct{})
		for loc := range *p.Board {
			occupied[loc] = struct{}{}
		}

		targetSquare, placed := p.RandomOpenSquare(occupied)
		if !placed {
			var eligible []Location
			for loc := range p.Squares {
				eligible = append(eligible, loc)
			}
			chosenIndex := rand.Intn(len(eligible))
			targetSquare = eligible[chosenIndex]
		}
		board := *p.Board
		if _, occ := board[targetSquare]; occ {
			p.BoardToBench(targetSquare)
		}
		p.BenchToBoard(kingIdx, targetSquare)
	}
}

func findOpenSquareForBlack(p *PlayerPieces, globalOccupied map[Location]struct{}, rng *rand.Rand) (loc Location, placed bool) {
	localExclusions := make(map[Location]struct{})
	for globalLoc := range globalOccupied {
		localLoc := globalLoc.Transform()
		localExclusions[localLoc] = struct{}{}
	}

	localChosen, placed := p.RandomOpenSquare(localExclusions)
	if localChosen.File == 0 {
		return Location{}, false
	}
	return localChosen.Transform(), true
}

func findOpenSquareForWhite(p *PlayerPieces, globalOccupied map[Location]struct{}, rng *rand.Rand) (loc Location, placed bool) {
	return p.RandomOpenSquare(globalOccupied)
}
