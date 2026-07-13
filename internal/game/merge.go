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

	whiteBB := whiteProfile.BoardAndBench
	ensureKingIsDeployed(whiteBB)

	for loc, piece := range *whiteBB.Board {
		piece.Color = White
		globalBoard[loc] = piece
	}

	blackBB := blackProfile.BoardAndBench
	ensureKingIsDeployed(blackBB)

	transposedBlackBoard := make(map[Location]Piece)
	for loc, piece := range *blackBB.Board {
		piece.Color = Black
		globalBoardLoc := loc.Transform() // Transform local absolute ranks into global upper ranks
		transposedBlackBoard[globalBoardLoc] = piece
	}

	//Collision detections
	occupiedSquares := make(map[Location]struct{})
	for loc := range globalBoard {
		occupiedSquares[loc] = struct{}{}
	}

	for blackLoc, blackPiece := range transposedBlackBoard {
		if _, collision := globalBoard[blackLoc]; collision {
			if rng.Float64() < 0.5 {
				// White wins the coin flip: bump Black to an open spot in Black's board
				bumpLoc, placed := findOpenSquareForBlack(blackBB, occupiedSquares, rng)
				if placed {
					globalBoard[bumpLoc] = blackPiece
					occupiedSquares[bumpLoc] = struct{}{}
				}
				// If placed is false, Black's piece is intentionally omitted/deleted from the board
			} else {
				// Black wins the coin flip: replace White and bump White to an open spot in White's territory
				whitePiece := globalBoard[blackLoc]
				globalBoard[blackLoc] = blackPiece

				bumpLoc, placed := findOpenSquareForWhite(whiteBB, occupiedSquares, rng)
				if placed {
					globalBoard[bumpLoc] = whitePiece
					occupiedSquares[bumpLoc] = struct{}{}
				}
				// If placed is false, White's piece is intentionally omitted/deleted from the board!
			}
		} else {
			globalBoard[blackLoc] = blackPiece
			occupiedSquares[blackLoc] = struct{}{}
		}
	}

	passAttempts := 0
	maxAttempts := 50 // Ensures we don't get into some form of resolution attempt loop

	for passAttempts < maxAttempts {
		legal, wk, bk, offending := LegalPosition(&globalBoard, White)
		if legal && wk && bk {
			return globalBoard, true, nil // Success! Perfect legal board configuration reached
		}

		//Force spawn missing king immediately
		if !wk {
			spawnKingOnRandomOpenSquare(&globalBoard, White, whiteBB, rng)
			maxAttempts++
			if maxAttempts >= 100 {
				maxAttempts = 100
			}
			continue
		}
		if !bk {
			spawnKingOnRandomOpenSquare(&globalBoard, Black, blackBB, rng)
			maxAttempts++
			if maxAttempts >= 100 {
				maxAttempts = 100
			}
			continue
		}

		//Loop through offending pieces and try to resolve them with movements to random open squares
		for _, targetLoc := range offending {
			piece, exists := globalBoard[targetLoc]
			if !exists {
				continue
			}

			// Generate exclusion matrix containing all currently occupied cells
			exclusions := make(map[Location]struct{})
			for loc := range globalBoard {
				exclusions[loc] = struct{}{}
			}

			// Forbid pawns from landing on back ranks during scatter passes
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
				// Swap and test legality downstream
				delete(globalBoard, targetLoc)
				globalBoard[candidateLoc] = piece

				testLegal, testWK, testBK, _ := LegalPosition(&globalBoard, White)
				if testLegal && testWK && testBK {
					break //We stablized the legality of the position
				}
				// Revert on validation failure
				delete(globalBoard, candidateLoc)
				globalBoard[targetLoc] = piece
			} else {
				//If placement is impossible, we remove the piece entirely.
				delete(globalBoard, targetLoc)
				break
			}
		}

		passAttempts++
	}

	//If we can not make a legal arrangement within the loop limit - we return a position with just kings
	finalLegal, finalWK, finalBK, _ := LegalPosition(&globalBoard, White)
	if !finalLegal || !finalWK || !finalBK {
		fallbackBoard := BoardState{
			Location{File: 5, Rank: 1}: {Type: King, Color: White},
			Location{File: 5, Rank: 8}: {Type: King, Color: Black},
		}
		return fallbackBoard, false, nil // Returns FALSE so the UI scene can render a specific warning log
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
