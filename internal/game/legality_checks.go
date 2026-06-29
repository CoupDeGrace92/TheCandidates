package game

// Pawns on back row
func LegalPosition(b *BoardState, activeColor PieceColor) (legal, wk, bk bool, loc []Location) {
	var whiteKingLoc, blackKingLoc Location
	whiteKingCount, blackKingCount := 0, 0
	legal, wk, bk = true, true, true

	var offending []Location

	if b == nil {
		legal = false
	}

	for loc, piece := range *b {
		switch piece.Type {
		case King:
			if piece.Color == White {
				whiteKingCount++
				if whiteKingCount > 1 {
					legal = false
					offending = append(offending, loc)
				}
				whiteKingLoc = loc
			} else {
				blackKingCount++
				if blackKingCount > 1 {
					legal = false
					offending = append(offending, loc)
				}
				blackKingLoc = loc
			}
		case Pawn:
			if loc.Rank == 1 || loc.Rank == 8 {
				legal = false
				offending = append(offending, loc)
			}
		}
	}
	if whiteKingCount != 1 {
		legal, wk = false, false
	}
	if blackKingCount != 1 {
		legal, bk = false, false
	}

	var inactiveKingLoc Location

	if activeColor == White {
		inactiveKingLoc = blackKingLoc
	} else {
		inactiveKingLoc = whiteKingLoc
	}

	if IsSquareAttacked(*b, inactiveKingLoc, activeColor) {
		legal = false
		offending = append(offending, inactiveKingLoc)
	}

	return legal, wk, bk, offending
}

func IsSquareAttacked(b BoardState, target Location, attackerColor PieceColor) bool {
	return countAttackers(b, target, attackerColor) > 0
}

// This structure existed when I thought stockfish would not take a position with 3+ pieces placing the king in check
func countAttackers(b BoardState, target Location, attackerColor PieceColor) int {
	attackers := 0

	//Knights
	knightOffsets := []struct{ f, r int }{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
	for _, o := range knightOffsets {
		loc := Location{File: target.File + o.f, Rank: target.Rank + o.r}
		if p, ok := b[loc]; ok && p.Type == Knight && p.Color == attackerColor {
			attackers++
		}
	}

	//Bishop, Rook, Queen vectors
	directions := []struct{ df, dr int }{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	for _, d := range directions {
		f, r := target.File+d.df, target.Rank+d.dr
		for f >= 1 && f <= 8 && r >= 1 && r <= 8 {
			loc := Location{File: f, Rank: r}
			if p, ok := b[loc]; ok {
				if p.Color == attackerColor {
					isDiagonal := d.df != 0 && d.dr != 0
					if p.Type == Queen || (isDiagonal && p.Type == Bishop) || (!isDiagonal && p.Type == Rook) {
						attackers++
					}
				}
				break
			}
			f += d.df
			r += d.dr
		}
	}

	//Pawns
	pawnRankOffset := -1
	if attackerColor == White {
		pawnRankOffset = 1
	}
	pawnFiles := []int{target.File - 1, target.File + 1}
	for _, f := range pawnFiles {
		loc := Location{File: f, Rank: target.Rank + pawnRankOffset}
		if p, ok := b[loc]; ok && p.Type == Pawn && p.Color == attackerColor {
			attackers++
		}
	}

	//TOUCHING KINGS!!!!
	for df := -1; df <= 1; df++ {
		for dr := -1; dr <= 1; dr++ {
			if df == 0 && dr == 0 {
				continue
			}
			loc := Location{File: target.File + df, Rank: target.Rank + dr}
			if p, ok := b[loc]; ok && p.Type == King && p.Color == attackerColor {
				attackers++
			}
		}
	}

	return attackers
}
