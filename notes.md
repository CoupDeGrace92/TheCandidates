# Design Notes
## The Candidates - Project Directory Architecture

This document maps out the Go-based backend and Ebitengine directory structure for **The Candidates**, an auto-chess strategy game leveraging Stockfish for combat simulations.

```text
the-candidates/
│
├── cmd/
│   └── game/
│       └── main.go          # Main entry point (initializes Ebitengine config and runs the application loop)
│
├── internal/
│   ├── ai/                  # Opponent AI logic: dictates CPU drafting strategy and shop behaviors
│   ├── campaign/            # Progression structures: unlocks, narrative milestones, and historical campaign maps
│   ├── draft/               # Drafting mechanics: shop pools, gold economy, and bench-to-board logistics
│   ├── engine/              # Stockfish harness: manages OS background processes, text pipes, and UCI streams
│   ├── game/                # Core domain state: shared structs (Piece, BoardSquare) and custom FEN string builders
│   ├── scene/               # Ebitengine UI loops: interface and layout switches (MenuScene, ShopScene, BattleScene)
│   └── tournament/          # Competitive formats: brackets, Swiss pairing systems, and leaderboard tables
│
├── assets/                  # Compiled binary data via Go's native 'go:embed' directive
│   ├── fonts/               # UI typography (clean, minimalist font layouts)
│   └── images/              # Visual vectors (flat 2D chess pieces, boards, and strategic icon sprites)
│
├── go.mod                   # Project module dependencies
└── go.sum                   # Dependency checksum security verification
```

## Phase 1
   -  **Board State Representation (BSR)**:  Object oriented structs that maintain the location of pieces.
   -  **Move Execution**:  When stockfish says "bestmove e2e4", the go code moves the objects in the BSR.
   -  **FEN String Generation**: Convert the BSR into an FEN string to feed back into Stockfish.

Initially we want to offload the chess logic to Stockfish, this means we have to parse the info lines it sends.  
This takes the form of:  
> info depth 10 seldepth 12 score mate 1 nodes 8523 nps 121000 pv e7e8q
> bestmove e7e8q

This should be relatively easy to parse, but we will need to detect things like mate and draws.  For mate, that is easy.  
When stockfish declares mate x, we know there is going to be a mate.  Then we can just play the game out until it reads:
> score mate 0
or
> score mate -1

For draws - this is trickier.  There are three types of draws we feasibly want to track:
 - Threefold repitition
 - Stalemate
 - Long periods of both engines reading: 
> score cp 0
 - (And also technically): the 50 move rule

The third is an interesting case - we can call it draw offered/accepted.  We are going to have to put  
some guard rails on this particular case to make sure we do not get some odd draws.

One thing to note, our BSR is only for our own pieces and not the opponents pieces.  These will be concatenated to form  
an FEN string to feed into Stockfish.  However since this is only done once, lets introduce:  
 - **Partial Board State Representation (pBSR)**: This will just be the represenation of a single players pieces.

## Phase 2
Here we will build the board and import some piece image assets.  Eventually these will be customizable, so we want to  
pull them from a player config file.  We have two interfaces to build:  
 - The Evaluation Board
 - The Drafting Board

These will be done in ebitengine as long as the liscence allows.

## GAME IDEAS
We want to build in some longevity for people playing.  This means we have to have some sort of progression.  This takes the form of:
 - Custom characters going through a campaign
 - Unlockables (both for within a single campaign and for an account)
   - Piece/Board cosmetics (unlock piece styles associated with famous tournaments for winning them on x difficulty)
   - Players
   - Skills?
 - Multiplayer Rankings

Long term - if we do actually have something on our hands, we will want to monitize:  
 - Campaign/Era expansions  
 - Paid cosmetics
 - Gacha packs for characters/skills if we go that direction.  DONT WANT P2W, KEEP BALANCED IF POSSIBLE