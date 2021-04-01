# game\_solver

## Introduction

This is a simple game solving AI that I wrote to test some ideas about
look-ahead game solving. I wanted to see the performance differences caused by
the number of options available to the player per turn, the number of cores
available, and the number of turns of lookahead performed. This program is in no way
optimized for performance, and is just a testbed. The bottom line: if you keep the number
of options per turn around 4, you can play in near-realtime with up to 9 turns of lookahead.
With 8 options, about 5 turns of lookahead is practical. This program uses goroutines for
concurrency, but due to the short evaluation time of most positions, it tends not to use all cores
on a computer, and doesn't achieve huge speed gains (maybe 1.1X to 2X, depending on depth and move options). A game with a
more complex evaluation function would likely benefit more from parallelism.

## Compilation

Just run

```sh
go build ai.go
```

### Usage

```sh
./ai depth moves
```

where `depth` is the number of turns of lookahead, and `moves` is the number of available turns.

### About available moves

Up to 24 moves are available to the player. The player must have at least 4
moves in order to solve the game. The first four moves are single steps in the
cardinal directions, the next four are diagonal single steps, the next eight
are straight double steps, and the next eight are knights' moves. 

### How it works

This AI evaluates all possible positions a set number of moves in the future.
Its evaluation function takes into account, in order of priority, death due to
being stepped on by a monster, the number of coins collected, the speed of coin
collection, and the distance to the nearest uncollected coin. This seems to be
more than enough information for it to avoid most problems.  Note that the algorithm
does not evaluate distance using a dijkstra map or similar, but merely adds the X and Y
distance to determine coin distance.
