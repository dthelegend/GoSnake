package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Direction int

const (
	North Direction = Direction(iota)
	South
	East
	West
)

func (d Direction) Opposite() Direction {
	switch d {
	case North:
		return South
	case South:
		return North
	case East:
		return West
	case West:
		return East
	}
	panic("Invalid Direction")
}

var ErrUnknownDirection error = errors.New("unknown direction")
var ErrCollision error = errors.New("rules")

type Position struct {
	X int
	Y int
}

func (p1 *Position) Add(p2 Position) {
	p1.X += p2.X
	p1.Y += p2.Y
}

func (p1 *Position) Subtract(p2 Position) {
	p1.X -= p2.X
	p1.Y -= p2.Y
}

type Snake struct {
	Head *Body
	Heading Direction
	Position Position
}

type Body struct {
	Tail      *Body
	RelativePosition Position
}

func (s *Snake) Size() int {
	var length int = 0
	for currentBody := s.Head; currentBody != nil; currentBody = currentBody.Tail {
		length++
	}
	return length
}

func (s *Snake) SetHeading(d Direction) (Direction) {
	if(s.Heading != d.Opposite()) {
		s.Heading = d
	}
	return s.Heading
}

func (s *Snake) Grow() *Body {
	// Create a new tail
	var newHead *Body = &Body{s.Head, Position{0,0}}

	if s.Head != nil {
		switch s.Heading {
		case North:
			s.Head.RelativePosition.Y++
		case West:
			s.Head.RelativePosition.X++
		case South:
			s.Head.RelativePosition.Y--
		case East:
			s.Head.RelativePosition.X--
		}
		s.Position.Subtract(s.Head.RelativePosition)
	}

	// Replace old tail
	s.Head = newHead

	return newHead
}

func (s *Snake) Shrink() *Body {
	for currentBody := s.Head; currentBody != nil; currentBody = currentBody.Tail {
		if currentBody.Tail != nil && currentBody.Tail.Tail == nil {
			var removed *Body= currentBody.Tail
			currentBody.Tail = nil
			return removed
		}
	}
	return nil
}

func (s *Snake) CheckCollision(maxWidth int, maxHeight int, positions ...Position) bool {
	var head *Body = s.Head

	if head != nil && len(positions) == 0 {
		// Skip Head in collisions
		positions = append(positions, s.Position)
		head = head.Tail
	}

	var snakePosition Position = s.Position
	for ; head != nil; head = head.Tail {
		snakePosition.Add(head.RelativePosition)
		if snakePosition.X < 0 {
			snakePosition.X += maxWidth
		} else if snakePosition.X >= maxWidth {
			snakePosition.X -= maxWidth
		}
		if snakePosition.Y < 0 {
			snakePosition.Y += maxHeight
		} else if snakePosition.Y >= maxHeight {
			snakePosition.Y -= maxHeight
		}
		for _, colliderPosition := range positions {
			if colliderPosition == snakePosition {
				return true
			}
		}
	}

	return false
}

func (p *Position) Draw(screen tcell.Screen, style tcell.Style) {
	var r rune
	r, _,  _, _ = screen.GetContent(p.X * 2, p.Y)
	screen.SetContent(p.X * 2, p.Y, r, nil, style)
	r, _,  _, _ = screen.GetContent(p.X * 2 + 1, p.Y)
	screen.SetContent(p.X * 2 + 1, p.Y, r, nil, style)
}

func main() {
	var style tcell.Style = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	var snakeStyle tcell.Style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	var pelletStyle tcell.Style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorPurple)

	screen, screen_err := tcell.NewScreen()
	if screen_err != nil {
		panic(screen_err)
	}
	
	screen_init_err := screen.Init()
	if screen_init_err != nil {
		panic(screen_init_err)
	}
	defer screen.Fini()

	screen.SetStyle(style)

	var snake Snake = Snake{Heading: South}
	var newDirection Direction = snake.Heading;
	var pellet *Position = nil
	var windowWidth int
	var windowHeight int
	
	// Event Loop
	go func() {
		// ensure that Fini is called
		defer screen.Fini()

		for {
			// Gather Input
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				windowWidth, windowHeight= ev.Size()
				windowWidth /= 2 
				if(pellet != nil && (pellet.X < 0 || pellet.X >= windowWidth || pellet.Y < 0 || pellet.Y >= windowHeight)) {
					pellet = nil
				}
				screen.Sync()
				continue
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlC {
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

	// Render and Game Loop
	// TODO : Split this into a Render Loop and a Game Loop
	for frame := 0; ; frame++ {
		// Clear Screen For Drawing
		screen.Clear()

		// Draw Static elements
		for i, r := range fmt.Sprintf("Score: %d", snake.Size()) {
			screen.SetContent(i,0, r, nil, style)
		}
		for i, r := range fmt.Sprintf("Frame: %d", frame) {
			screen.SetContent(i,1, r, nil, style)
		}
		for i, r := range fmt.Sprintf("Position: (%d, %d)", snake.Position.X, snake.Position.Y) {
			screen.SetContent(i,2, r, nil, style)
		}
		if pellet != nil {
			for i, r := range fmt.Sprintf("Pellet: (%d, %d)", pellet.X, pellet.Y) {
				screen.SetContent(i,3, r, nil, style)
			}
		} else {
			for i, r := range "Pellet: Nil" {
				screen.SetContent(i,3, r, nil, style)
			}
		}
		for i, r := range fmt.Sprintf("Window Size: (%d, %d)", windowWidth, windowHeight) {
			screen.SetContent(i,4, r, nil, style)
		}
		for i, r := range fmt.Sprintf("Direction: %d", newDirection) {
			screen.SetContent(i,5, r, nil, style)
		}

		// Move and Render Snake
		// Set Snake Direction
		snake.SetHeading(newDirection)

		// Grow Snake
		snake.Grow()
		if snake.Position.X < 0 {
			snake.Position.X += windowWidth
		} else if snake.Position.X >= windowWidth {
			snake.Position.X -= windowWidth
		}
		if snake.Position.Y < 0 {
			snake.Position.Y += windowHeight
		} else if snake.Position.Y >= windowHeight {
			snake.Position.Y -= windowHeight
		}

		// Shrink Snake
		if pellet != nil && snake.Position == *pellet {
			pellet = nil
		} else {
			snake.Shrink()
		}

		// Check for collisions
		if snake.CheckCollision(windowWidth, windowHeight) {
			panic("rules")
		}

		// Draw Snake
		{
			var currentPosition Position = snake.Position
			for currentBody := snake.Head; currentBody != nil; currentBody = currentBody.Tail {
				currentPosition.Add(currentBody.RelativePosition)
				if currentPosition.X < 0 {
					currentPosition.X += windowWidth
				} else if currentPosition.X >= windowWidth {
					currentPosition.X -= windowWidth
				}
				if currentPosition.Y < 0 {
					currentPosition.Y += windowHeight
				} else if currentPosition.Y >= windowHeight {
					currentPosition.Y -= windowHeight
				}
				currentPosition.Draw(screen, snakeStyle)
			}
		}

		// Draw Pellet
		if pellet != nil {
			pellet.Draw(screen, pelletStyle)
		} else {
			for inSnake := true; inSnake; {
				inSnake = false
				pellet = &Position{rand.Intn(windowWidth), rand.Intn(windowHeight)}
				if snake.CheckCollision(windowWidth, windowHeight, *pellet) {
					inSnake = true
				}
			}
		}

		// Update screen
		screen.Show()

		// Sleep
		time.Sleep(time.Second / 6)
	}
}
