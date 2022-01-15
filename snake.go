package main

import (
	"fmt"
	"math/rand"
	"time"

	term "github.com/nsf/termbox-go"
)

const (
	up    = 0
	right = 1
	down  = 2
	left  = 3
	none  = 4
)

type Board struct {
	SizeX, SizeY   int
	SnakeX, SnakeY []int
	SnakeLen       int
	AppleX, AppleY []int
	AppleAmount    int
	Score          int
}

func CreateBoard(x, y, apples int) (b Board) {
	b.SizeX = x
	b.SizeY = y
	b.SnakeLen = 1
	b.AppleAmount = apples
	b.SnakeX = make([]int, 1, x*y)
	b.SnakeY = make([]int, 1, x*y)
	b.SnakeX[0] = x / 2
	b.SnakeY[0] = y / 2
	b.AppleX = make([]int, apples)
	b.AppleY = make([]int, apples)
	for i := 0; i < apples; i++ {
		b.AppleX[i] = rand.Intn(x - 1)
		b.AppleY[i] = rand.Intn(y - 1)
	}
	b.Score = 0
	return b
}

func (b Board) UpdateBoard(input uint8) (dead bool) {
	dead = false
	prevSnakeX := b.SnakeX[0] // Previous tail position, set now but used in iterations later
	prevSnakeY := b.SnakeY[0]
	switch input { // input handling, obviously
	case up:
		if b.SnakeLen > 1 && b.SnakeY[1] == b.SnakeY[0]+1 { // Make sure the snake can't reverse
			b.SnakeY[0] -= 1
		} else {
			b.SnakeY[0] += 1
		}
	case right:
		if b.SnakeLen > 1 && b.SnakeX[1] == b.SnakeX[0]+1 {
			b.SnakeX[0] -= 1
		} else {
			b.SnakeX[0] += 1
		}
	case down:
		if b.SnakeLen > 1 && b.SnakeY[1] == b.SnakeY[0]-1 {
			b.SnakeY[0] += 1
		} else {
			b.SnakeY[0] -= 1
		}
	case left:
		if b.SnakeLen > 1 && b.SnakeX[1] == b.SnakeX[0]-1 {
			b.SnakeX[0] += 1
		} else {
			b.SnakeX[0] -= 1
		}
	}

	if b.SnakeX[0] >= b.SizeX || b.SnakeX[0] < 0 || b.SnakeY[0] >= b.SizeY || b.SnakeY[0] < 0 { // wall collision
		dead = true
		return
	}

	for i := 0; i < b.AppleAmount; i++ { // apple collision
		if b.SnakeX[0] == b.AppleX[i] && b.SnakeY[0] == b.AppleY[i] {
			b.Score += 1
			b.AppleX[i] = rand.Intn(b.SizeX - 1)
			b.AppleY[i] = rand.Intn(b.SizeY - 1)
		}
	}

	for i := 1; i < b.SnakeLen; i++ { // snek collision and redrawing
		// look I can swap two integers without a third, aren't I cool
		prevSnakeX += b.SnakeX[i]
		prevSnakeY += b.SnakeY[i]
		b.SnakeX[i] = prevSnakeX - b.SnakeX[i]
		b.SnakeY[i] = prevSnakeY - b.SnakeY[i]
		prevSnakeX -= b.SnakeX[i]
		prevSnakeY -= b.SnakeY[i]
		if b.SnakeX[i] == b.SnakeX[0] && b.SnakeY[i] == b.SnakeY[0] { // snek collision
			dead = true
			return
		}
		if b.Score+1 > b.SnakeLen { // add new snek segment
			b.SnakeLen = b.Score + 1
			b.SnakeX = b.SnakeX[:b.SnakeLen]
			b.SnakeY = b.SnakeY[:b.SnakeLen]
			b.SnakeX[b.SnakeLen] = prevSnakeX
			b.SnakeY[b.SnakeLen] = prevSnakeY
		}
	}
	return
}

func (b Board) String() string {
	x, y := b.SizeX, b.SizeY         // not dealing with writing out b.Size* every time
	ret := make([]byte, (x+3)*(y+2)) // A one-dimensional array of the map. Has 1-byte borders, so uses x+2 and y+2. The additional +1 in the x is for \n.
	// start by drawing the top line, with the score display
	scoreDisplay := " Score: " + string(b.Score) + " "
	scoreRelativeIndex := (x+2)/2 - len(scoreDisplay)/2 // finds the length of the border before it hits the centered score display

	// set all bytes in the array to space. Very time consuming linear algorithm, and half of them will get overwritten anyway. Oh well.
	for i := 0; i < (x+3)*(y+2); i++ {
		ret[i] = ' '
	}

	for i := 0; i < x+2; i++ { // draw top border
		if i-scoreRelativeIndex < 0 || i >= scoreRelativeIndex+len(scoreDisplay) {
			ret[i] = '#' // border sign
		} else {
			ret[i] = scoreDisplay[i-scoreRelativeIndex]
		}
	}

	for i := 0; i < y+2; i++ { // draw left border
		ret[i*(x+3)] = '#' // left
	}
	for i := 0; i < x+3; i++ { // draw bottom border
		ret[(x+3)*(y+1)+i] = '#'
	}
	for i := 0; i < y+2; i++ { // draw right border and newline
		ret[(i+1)*(x+3)-2] = '#'  // right
		ret[(i+1)*(x+3)-1] = '\n' // newline
	}

	// render le snek
	for i := 0; i < b.SnakeLen; i++ {
		ret[(x+3)*(y+1-b.SnakeY[i])+b.SnakeX[i]+1] = '@' // snek is rendered as @s
	}

	// render le apples. This code may look familiar
	for i := 0; i < b.AppleAmount; i++ {
		ret[(x+3)*(y+1-b.AppleY[i])+b.AppleX[i]+1] = 'O' // apples is rendered as Os
	}

	return string(ret)
}

func HandleIStream(input *uint8, running *bool) { // Handle input stream (arrow keys)
	for *running {
		switch ev := term.PollEvent(); ev.Key {
		case term.KeyEsc:
			if *input != none { // prevent disabling the input before the game starts
				*running = false
			}
		case term.KeyArrowUp:
			*input = up
		case term.KeyArrowRight:
			*input = right
		case term.KeyArrowDown:
			*input = down
		case term.KeyArrowLeft:
			*input = left
		default:
			continue // idk
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano()) // seed random
	board := CreateBoard(40, 40, 5)
	dead := false
	kill := new(bool)
	input := new(uint8)

	term.Init() // init term

	*kill = true // default values
	*input = none

	go HandleIStream(input, kill)

	fmt.Print(board)
	for *input == none { // wait on input
		time.Sleep(500 * time.Millisecond)
	}

	for dead == false { // begin game
		time.Sleep(500 * time.Millisecond)
		term.Sync()
		dead = board.UpdateBoard(*input)
		fmt.Print(board)
		if *kill == false {
			dead = true
		}
	}

	*kill = false                                    // stop input
	term.Close()                                     // close term
	fmt.Println("Your final score was", board.Score) // Print score
}
