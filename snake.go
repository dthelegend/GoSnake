package main

import (
	"github.com/gdamore/tcell/v2"
	"time"
)

type Direction int

const (
	North Direction = Direction(iota)
	South
	East
	West
)

type Position struct {
	X int
	Y int
}

type Snake struct {
	Tail      *Snake
	Direction Direction
	Position  *Position
}

func (s *Snake) Move(d Direction) {
	// Move Tail
	if s.Tail != nil {
		s.Tail.Move(s.Direction)
	}
	s.Direction = d;

	// Modify Position
	switch s.Direction {
	case North:
		s.Position.Y -= 1
	case South:
		s.Position.Y += 1
	case East:
		s.Position.X += 1
	case West:
		s.Position.X -= 1
	default:
		panic("Unknown Direction!")
	}
}

func main() {
	var snake *Snake = &Snake{nil, South, &Position{0,0}}
	var style tcell.Style = tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorGreen)

	var newDirection Direction = snake.Direction;

	screen, screen_err := tcell.NewScreen()
	if screen_err != nil {
		panic(screen_err)
	}
	
	screen_init_err := screen.Init()
	if screen_init_err != nil {
		panic(screen_init_err)
	}

	screen.SetStyle(style)
	screen.Clear()

	go func() {
		for {
			// Gather Input
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				screen.Sync()
				continue
			case *tcell.EventKey:
				if(ev.Key() == tcell.KeyCtrlC){
					panic("Ok Boss")
				}
				switch ev.Rune() {
				case 'W':
					fallthrough
				case 'w':
					newDirection = North
				case 'A':
					fallthrough
				case 'a':
					newDirection = West
				case 'S':
					fallthrough
				case 's':
					newDirection = South
				case 'D':
					fallthrough
				case 'd':
					newDirection = East
				}
			}

		}
	}()

	for{
		// Render
		for currentSnake := snake; currentSnake != nil; currentSnake = snake.Tail {
			screen.SetContent(currentSnake.Position.X, currentSnake.Position.Y, '*', nil, style)
		}
		// Move Snake
		snake.Move(newDirection)
		// Update screen
		screen.Show()

		time.Sleep(time.Second)
	}
}