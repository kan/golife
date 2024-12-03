package game

import (
	"sync"

	"github.com/kan/golife/parser"
)

type Game struct {
	world  GameBoard
	shadow GameBoard
	rule   *parser.Rule
	run    bool
	size   int
	mu     sync.Mutex
}

type GameBoard [][]uint

func NewGame(size int, rule *parser.Rule) *Game {
	world := make([][]uint, size)
	shadow := make([][]uint, size)
	for i := 0; i < size; i++ {
		world[i] = make([]uint, size)
		shadow[i] = make([]uint, size)
	}

	return &Game{
		world:  world,
		shadow: shadow,
		rule:   rule,
		run:    false,
		size:   size,
	}
}

func (g *Game) StartStop() {
	g.run = !g.run
}

func (g *Game) Run() bool {
	return g.run
}

func (g *Game) Size() int {
	return g.size
}

func (g *Game) GetWorld(x, y int) bool {
	return g.world[x][y] == 1
}

func (g *Game) SetWorld(w [][]uint) {
	g.world = w
}

func (g *Game) Toggle(x, y int) {
	if x >= g.size || y >= g.size {
		return
	}

	if g.GetWorld(x, y) {
		g.world[x][y] = 0
	} else {
		g.world[x][y] = 1
	}
}

func (g *Game) ApplyBoard(w int, f func(x, y int)) {
	var wg sync.WaitGroup
	jobs := make(chan [2]int, g.size*g.size)

	for i := 0; i < w; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				f(job[0], job[1])
			}
		}()
	}

	for y := 0; y < g.size; y++ {
		for x := 0; x < g.size; x++ {
			jobs <- [2]int{x, y}
		}
	}
	close(jobs)

	wg.Wait()
}

func (g *Game) Step(w int) {
	for y := 0; y < g.size; y++ {
		for x := 0; x < g.size; x++ {
			g.shadow[x][y] = 0
		}
	}

	g.ApplyBoard(w, func(x, y int) {
		g.stepBoard(x, y)
	})

	g.mu.Lock()
	g.world, g.shadow = g.shadow, g.world
	g.mu.Unlock()
}

func (g *Game) stepBoard(x, y int) {
	sw := uint(0)
	if x-1 >= 0 && y-1 >= 0 {
		sw += g.world[x-1][y-1]
	}
	if y-1 >= 0 {
		sw += g.world[x][y-1]
	}
	if x+1 < g.size && y-1 >= 0 {
		sw += g.world[x+1][y-1]
	}
	if x-1 >= 0 {
		sw += g.world[x-1][y]
	}
	if x+1 < g.size {
		sw += g.world[x+1][y]
	}
	if x-1 >= 0 && y+1 < g.size {
		sw += g.world[x-1][y+1]
	}
	if y+1 < g.size {
		sw += g.world[x][y+1]
	}
	if x+1 < g.size && y+1 < g.size {
		sw += g.world[x+1][y+1]
	}

	if g.world[x][y] == 1 {
		if g.rule.IsSurvival(sw) {
			g.shadow[x][y] = 1
		} else {
			g.shadow[x][y] = 0
		}
	} else {
		if g.rule.IsBirth(sw) {
			g.shadow[x][y] = 1
		} else {
			g.shadow[x][y] = 0
		}
	}
}

func (g *Game) Save(filename string) error {
	return parser.SaveRLE(filename, g.world, g.rule)
}
