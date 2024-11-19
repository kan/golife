package main

import (
	"runtime"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

const SIZE = 20

var world = [][]uint{}
var shadowWorld = [][]uint{}
var run = false
var interval = 500 * time.Millisecond

func genWorld() [][]uint {
	w := [][]uint{}
	for i := 0; i < SIZE; i++ {
		row := make([]uint, SIZE)
		w = append(w, row)
	}

	return w
}

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	w := runtime.NumCPU()

	if err := screen.Init(); err != nil {
		panic(err)
	}
	defer screen.Fini()

	screen.EnableMouse()
	defer screen.DisableMouse()

	world = genWorld()
	shadowWorld = genWorld()

	eventLoop(screen, w)
}

func eventLoop(screen tcell.Screen, w int) {
	eventCh := make(chan tcell.Event)

	go func() {
		for {
			eventCh <- screen.PollEvent()
		}
	}()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case ev := <-eventCh:
			switch e := ev.(type) {
			case *tcell.EventKey:
				if e.Key() == tcell.KeyEscape {
					return
				} else if e.Key() == tcell.KeyCtrlS {
					run = !run
				} else if e.Rune() == '+' {
					if interval > 100*time.Millisecond {
						interval -= 100 * time.Millisecond
						ticker.Reset(interval)
					}
				} else if e.Rune() == '-' {
					interval += 100 * time.Millisecond
					ticker.Reset(interval)
				}
			case *tcell.EventMouse:
				x, y := e.Position()
				btn := e.Buttons()

				if int(x/2) >= SIZE || y >= SIZE {
					continue
				}

				if btn&tcell.Button1 != 0 {
					if world[int(x/2)][y] == 0 {
						world[int(x/2)][y] = 1
					} else {
						world[int(x/2)][y] = 0
					}
				}
				showWorld(screen, w)
			case *tcell.EventResize:
				screen.Sync()
			}

		case <-ticker.C:
			if run {
				lifeWorld(w)
			}
			showWorld(screen, w)
		}
	}
}

func showWorld(screen tcell.Screen, w int) {
	screen.Clear()

	applyWorld(screen, w, dispCell)

	screen.Show()
}

func lifeWorld(w int) {
	for y := 0; y < SIZE; y++ {
		for x := 0; x < SIZE; x++ {
			shadowWorld[x][y] = 0
		}
	}
	applyWorld(nil, w, lifeCell)

	world, shadowWorld = shadowWorld, world
}

func applyWorld(screen tcell.Screen, w int, f func(tcell.Screen, int, int)) {
	var wg sync.WaitGroup
	jobs := make(chan [2]int, SIZE*SIZE)

	for i := 0; i < w; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				f(screen, job[0], job[1])
			}
		}()
	}

	for y, row := range world {
		for x := range row {
			jobs <- [2]int{x, y}
		}
	}
	close(jobs)

	wg.Wait()
}

var gs = tcell.StyleDefault.Foreground(tcell.ColorGreen)
var bs = tcell.StyleDefault.Foreground(tcell.ColorBlack)

func dispCell(screen tcell.Screen, x, y int) {
	if world[x][y] == 0 {
		screen.SetContent(x*2, y, rune('■'), nil, bs)
	} else {
		screen.SetContent(x*2, y, rune('■'), nil, gs)
	}
}

/*
セルの生死は次のルールに従う。

誕生
死んでいるセルに隣接する生きたセルがちょうど3つあれば、次の世代が誕生する。
生存
生きているセルに隣接する生きたセルが2つか3つならば、次の世代でも生存する。
過疎
生きているセルに隣接する生きたセルが1つ以下ならば、過疎により死滅する。
過密
生きているセルに隣接する生きたセルが4つ以上ならば、過密により死滅する。
*/
func lifeCell(screen tcell.Screen, x, y int) {
	sw := uint(0)
	if x-1 >= 0 && y-1 >= 0 {
		sw += world[x-1][y-1]
	}
	if y-1 >= 0 {
		sw += world[x][y-1]
	}
	if x+1 < SIZE && y-1 >= 0 {
		sw += world[x+1][y-1]
	}
	if x-1 >= 0 {
		sw += world[x-1][y]
	}
	if x+1 < SIZE {
		sw += world[x+1][y]
	}
	if x-1 >= 0 && y+1 < SIZE {
		sw += world[x-1][y+1]
	}
	if y+1 < SIZE {
		sw += world[x][y+1]
	}
	if x+1 < SIZE && y+1 < SIZE {
		sw += world[x+1][y+1]
	}

	if sw < 2 {
		shadowWorld[x][y] = 0
	} else if sw == 3 {
		shadowWorld[x][y] = 1
	} else if sw > 3 {
		shadowWorld[x][y] = 0
	} else {
		shadowWorld[x][y] = world[x][y]
	}
}
