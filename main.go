package main

import (
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

const SIZE = 20

var world = [][]uint{}
var newWorld = [][]uint{}
var run = false

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

	if err := screen.Init(); err != nil {
		panic(err)
	}
	defer screen.Fini()

	screen.EnableMouse()
	defer screen.DisableMouse()

	world = genWorld()

	eventLoop(screen)
}

func eventLoop(screen tcell.Screen) {
	eventCh := make(chan tcell.Event)

	go func() {
		for {
			eventCh <- screen.PollEvent()
		}
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
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
				showWorld(screen)
			case *tcell.EventResize:
				screen.Sync()
			}

		case <-ticker.C:
			if run {
				lifeWorld()
			}
			showWorld(screen)
		}
	}
}

func showWorld(screen tcell.Screen) {
	screen.Clear()

	applyWorld(screen, dispCell)

	screen.Show()
}

func lifeWorld() {
	newWorld = genWorld()
	applyWorld(nil, lifeCell)

	world = newWorld
}

func applyWorld(screen tcell.Screen, f func(tcell.Screen, int, int, *sync.WaitGroup)) {
	var wg sync.WaitGroup

	for y, row := range world {
		for x := range row {
			wg.Add(1)
			go f(screen, x, y, &wg)
		}
	}

	wg.Wait()
}

var gs = tcell.StyleDefault.Foreground(tcell.ColorGreen)
var bs = tcell.StyleDefault.Foreground(tcell.ColorBlack)

func dispCell(screen tcell.Screen, x, y int, wg *sync.WaitGroup) {
	defer wg.Done()

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
func lifeCell(screen tcell.Screen, x, y int, wg *sync.WaitGroup) {
	defer wg.Done()

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
		newWorld[x][y] = 0
	} else if sw == 3 {
		newWorld[x][y] = 1
	} else if sw > 3 {
		newWorld[x][y] = 0
	} else {
		newWorld[x][y] = world[x][y]
	}
}
