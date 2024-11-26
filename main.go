package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kan/golife/game"
	"github.com/kan/golife/parser"
)

var lifeGame *game.Game
var interval = 500 * time.Millisecond

func main() {
	size := flag.Int("size", 20, "盤面のサイズ")
	file := flag.String("file", "", "盤面のパターンファイル(RLE形式)")
	flag.Parse()

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

	var board [][]uint
	var rule *parser.Rule

	if file != nil && *file != "" {
		board, rule, err = parser.LoadRLE(*file, *size)
		if err != nil {
			panic(err)
		}
	} else {
		rule, err = parser.ParseRule("B3/S23")
		if err != nil {
			panic(err)
		}
	}

	lifeGame = game.NewGame(*size, rule)
	if len(board) > 0 {
		lifeGame.SetWorld(board)
	}

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
					lifeGame.StartStop()
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

				if btn&tcell.Button1 != 0 {
					lifeGame.Toggle(int(x/2), y)
				}
				showWorld(screen, w)
			case *tcell.EventResize:
				screen.Sync()
			}

		case <-ticker.C:
			if lifeGame.Run() {
				lifeGame.Step(w)
			}
			showWorld(screen, w)
		}
	}
}

var gs = tcell.StyleDefault.Foreground(tcell.ColorGreen)
var bs = tcell.StyleDefault.Foreground(tcell.ColorBlack)

func showWorld(screen tcell.Screen, w int) {
	screen.Clear()

	lifeGame.ApplyBoard(w, func(x, y int) {
		if lifeGame.GetWorld(x, y) {
			screen.SetContent(x*2, y, rune('■'), nil, gs)
		} else {
			screen.SetContent(x*2, y, rune('■'), nil, bs)
		}
	})

	screen.Show()
}
