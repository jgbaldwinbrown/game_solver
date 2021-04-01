package main

import (
    "fmt"
    "os"
    "strconv"
    "sync"
)

type pos struct {
    x int
    y int
}

type vec pos

type player struct {
    p pos
    alive bool
}

type coin struct {
    p pos
    available bool
}

type enemy struct {
    p pos
    move_dir vec
    move_count int
    alive bool
}

type state struct {
    player player
    coins []coin
    enemies []enemy
    coins_collected int
    turn int
    time_bonus int
}

type score_return struct {
    score float64
    state state
    plan [15]vec
}

var move_vecs [4]vec

func (s state) alive_score() float64 {
    if s.player.alive {
        return 1000.0
    } else {
        return 0.0
    }
}

func (s state) coin_score() float64 {
    return float64(s.coins_collected) + 0.1 * float64(s.time_bonus)
}

func min(i1 int, i2 int) int {
    if i1 < i2 {
        return i1
    }
    return i2
}

func max(i1 int, i2 int) int {
    if i1 > i2 {
        return i1
    }
    return i2
}

func abs(i int) int {
    if i < 0 {
        return -i
    }
    return i
}

func (p pos) move(v vec) pos {
    p.x += v.x
    p.y += v.y
    return p
}

func vec_dist(p1 pos, p2 pos) int {
    return abs(p1.x - p2.x) + abs(p1.y - p2.y)
}

func (s state) dist_score() float64 {
    var min_dist int = 1000000
    found_one := false
    for _, acoin := range s.coins {
        if acoin.available {
            found_one = true
            min_dist = min(min_dist, vec_dist(s.player.p, acoin.p))
        }
    }
    if !found_one {
        min_dist = 0
    }
    return -0.0001 * float64(min_dist)
}

func (s state) copy_state() state {
    var out state
    out.coins_collected = s.coins_collected
    out.player = s.player

    out.coins = make([]coin, len(s.coins))
    copy(out.coins, s.coins)
    out.enemies = make([]enemy, len(s.enemies))
    copy(out.enemies, s.enemies)
    out.turn = s.turn
    out.time_bonus = s.time_bonus

    return out
}

func (s state) update(v vec) state {
    n := s.copy_state()
    n.turn--
    n.player.p = s.player.p.move(v)
    for index, anenemy := range n.enemies {
        if anenemy.move_count > 0 {
            n.enemies[index].move_count--
        } else {
            n.enemies[index].move_count = 13
            n.enemies[index].move_dir.x *= -1
            n.enemies[index].move_dir.y *= -1
        }
        n.enemies[index].p = anenemy.p.move(anenemy.move_dir)
    }

    for index, acoin := range n.coins {
        if acoin.p == n.player.p && acoin.available {
            n.coins_collected++
            n.coins[index].available = false
            n.time_bonus += n.turn
        }
    }

    for _, anenemy := range n.enemies {
        if anenemy.p == n.player.p {
            n.player.alive = false
        }
    }

    return n
}
func (s state) score(depth int) (float64, state, [15]vec) {
    var return_scores [len(move_vecs)]score_return
    var my_score float64 = -1.0e25
    var my_state state
    var my_move vec
    var wg sync.WaitGroup
    var planned_moves [15]vec

    for i := 0; i < len(move_vecs); i++ {
        new_state := s.update(move_vecs[i])
        wg.Add(1)
        go new_state.score_parallel(&wg, depth - 1, planned_moves, &return_scores[i])
    }

    wg.Wait()

    for i := 0; i < len(move_vecs); i++ {
        this_return := return_scores[i]
        if my_score < this_return.score {
            my_score = this_return.score
            my_state = this_return.state
            my_move = move_vecs[i]
            planned_moves = this_return.plan
        }
    }
    planned_moves[depth] = my_move
    return my_score, my_state, planned_moves
}

func (s state) score_parallel(parent_wg *sync.WaitGroup, depth int, planned_moves [15]vec, my_return *score_return) {
    defer parent_wg.Done()

    var return_scores [len(move_vecs)]score_return
    var my_score float64 = -1.0e25
    var my_state state
    var my_move vec
    var wg sync.WaitGroup

    if depth > 0 {
        for i := 0; i < len(move_vecs); i++ {
            new_state := s.update(move_vecs[i])
            wg.Add(1)
            new_state.score_parallel(&wg, depth - 1, planned_moves, &return_scores[i])
        }
        wg.Wait()

        for i := 0; i < len(move_vecs); i++ {
            this_return := return_scores[i]
            if my_score < this_return.score {
                my_score = this_return.score
                my_state = this_return.state
                my_move = move_vecs[i]
                planned_moves = this_return.plan
            }
        }
    } else {
        my_score = s.alive_score() + s.coin_score() + s.dist_score()
        my_state = s
        my_move = vec{0,0}
        /*
        fmt.Printf("Scoring, depth = %v\n", depth)
        s.print_state()
        fmt.Printf("%v\n", s);
        fmt.Printf("Score: %v\n", my_score)
        */
    }
    planned_moves[depth] = my_move
    my_return.state = my_state
    my_return.plan = planned_moves
    my_return.score = my_score
}

func (s state) print_state() {
    print_matrix := make([]string, (40 * 20))
    for i := 0; i < (40 * 20); i++ {
        print_matrix[i] = " "
    }
    for _, c := range s.coins {
        print_matrix[c.p.y * 40 + c.p.x] = "$"
    }
    for _, e := range s.enemies {
        print_matrix[e.p.y * 40 + e.p.x] = "M"
    }
    if s.player.p.x >= 0 && s.player.p.y >= 0 {
        print_matrix[s.player.p.y * 40 + s.player.p.x] = "@"
    }

    fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
    for y := 0; y<20; y++ {
        fmt.Printf("|")
        for x := 0; x<40; x++ {
            fmt.Printf(" %s", print_matrix[y*40 + x])
        }
        fmt.Printf("|\n")
    }
    fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
}

func init_state() state {
    s := state{
        player: player {
            p: pos{x: 1, y: 10},
            alive: true,
        },
        coins: make([]coin, 0),
        enemies: make([]enemy, 0),
        coins_collected: 0,
    }
    s.coins = append(s.coins, coin{p: pos{x: 18, y: 10}, available: true})
    s.coins = append(s.coins, coin{p: pos{x: 19, y: 10}, available: true})
    s.coins = append(s.coins, coin{p: pos{x: 10, y: 5}, available: true})
    s.coins = append(s.coins, coin{p: pos{x: 10, y: 10}, available: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 8, y: 4}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 9, y: 4}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 10, y: 4}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 11, y: 4}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 12, y: 4}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 13, y: 4}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 8, y: 5}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 9, y: 5}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 10, y: 5}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 11, y: 5}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 12, y: 5}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.enemies = append(s.enemies, enemy{p: pos{x: 13, y: 5}, move_dir: vec{x: 0, y: 1}, move_count: 13, alive: true})
    s.turn = 100000
    s.time_bonus = 0
    return s
}

func (s state) alive() bool {
    return s.player.alive
}

func (s state) coins_remain() bool {
    for _, acoin := range s.coins {
        if acoin.available {
            return true
        }
    }
    return false
}

func (s state) take_move(depth int) state {
    chosen_move := vec{x: 0, y: 0}
    move_score := float64(-1.0e20)
    // var planned_moves [15]vec
    // var ideal_state state
    // fmt.Printf("Testing moves\n")
    for _, move := range move_vecs {
        // fmt.Printf("Testing move %v\n", move)
        new_score, _, _ := s.update(move).score(depth)
        if new_score >= move_score {
            chosen_move = move
            move_score = new_score
            // planned_moves = aplan
        }
        // fmt.Printf("Possible move: %v\n", move)
        // fmt.Printf("Possible score: %v\n", new_score)
        // fmt.Printf("Possible planned moves: %v\n", aplan)
    }
    // fmt.Printf("Done testing moves\n")
    // fmt.Printf("Taking move; best end state:\n", depth)
    // ideal_state.print_state()
    // fmt.Printf("%v\n", ideal_state);
    // fmt.Printf("Score: %v\n", move_score)
    // fmt.Printf("Chosen move: %v\n", chosen_move)
    // fmt.Printf("Planned moves: %v\n", planned_moves)
    // fmt.Printf("Done taking move.\n")
    return s.update(chosen_move)
}

func main() {
    move_vecs = [4]vec{
        vec {x: 1, y: 0},
        vec {x: -1, y: 0},
        vec {x: 0, y: 1},
        vec {x: 0, y: -1},
    }
    depth, err := strconv.Atoi(os.Args[1])
    if err != nil {
        panic(err)
    }

    // fmt.Printf("Current state:\n")
    my_state := init_state()
    my_state.print_state()
    // fmt.Printf("Done with current state.\n")
    // fmt.Printf("%v\n", my_state)
    for my_state.alive() && my_state.coins_remain() {
        my_state = my_state.take_move(depth)
        // fmt.Printf("Current state:\n")
        my_state.print_state()
        // fmt.Printf("%v\n", my_state)
        // fmt.Printf("Done with current state.\n")
    }
}
