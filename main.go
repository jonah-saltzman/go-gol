package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/mattn/go-tty"
)

type board = [][]int

type game struct {
	board  *board
	steps  int
	width  int
	height int
}

type position [2]int

func newGame(width, height int) *game {
	rand.Seed(time.Now().UnixNano())
	b := make(board, height)
	for i := range b {
		b[i] = make([]int, width)
		for j := range b[i] {
			b[i][j] = rand.Intn(2)
		}
	}

	game := &game{
		board:  &b,
		steps:  0,
		width:  width,
		height: height,
	}

	go game.listenForKeyPress()

	return game
}

func (g *game) getValue(pos position) int {
	b := g.board
	if pos[0] < 1 || pos[0] >= len(*b) {
		return 0
	}
	if pos[1] < 1 || pos[1] >= len((*b)[pos[0]]) {
		return 0
	}
	return (*b)[pos[0]][pos[1]]
}

func (g *game) listenForKeyPress() {
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()

	for {
		_, err := tty.ReadRune()
		if err != nil {
			panic(err)
		}

	}
}

func (g *game) beforeGame() {
	hideCursor()

	// handle CTRL C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			g.over()
		}
	}()
}

func (g *game) countLive() int {
	count := 0
	for i := range *g.board {
		for j := range (*g.board)[i] {
			if (*g.board)[i][j] == 1 {
				count++
			}
		}
	}
	return count
}

func (g *game) over() {
	clear()
	showCursor()

	moveCursor(position{1, 1})
	draw("game ended after " + strconv.Itoa(g.steps) + " steps\nlive cells: " + strconv.Itoa(g.countLive()) + "\n")

	render()

	os.Exit(0)
}

func (g *game) draw() {
	clear()
	maxX, _ := getSize()

	status := "live cells: " + strconv.Itoa(g.countLive())
	statusXPos := maxX/2 - len(status)/2

	moveCursor(position{statusXPos, 0})
	draw(status)

	for i := range *g.board {
		for j := range (*g.board)[i] {
			moveCursor(position{j + 1, i + 1})
			if (*g.board)[i][j] == 1 {
				draw("#")
			} else {
				draw(" ")
			}
		}
	}

	render()
	time.Sleep(time.Millisecond * 50)
}

func (g *game) liveNeighbors(p position) int {
	live := 0
	for i := p[0] - 1; i <= p[0]+1; i++ {
		for j := p[1] - 1; j <= p[1]+1; j++ {
			live += g.getValue(position{i, j})
		}
	}
	return live
}

func (g *game) copyBoard() board {
	b := make(board, g.height)
	for i := range b {
		b[i] = make([]int, g.width)
		for j := range b[i] {
			b[i][j] = (*g.board)[i][j]
		}
	}
	return b
}

func (g *game) tick() {
	g.board = g.nextBoard()
	g.draw()
	g.steps++
}

func (g *game) nextBoard() *board {
	newBoard := g.copyBoard()
	for i := range *g.board {
		for j := range (*g.board)[i] {
			neighbors := g.liveNeighbors(position{i, j})
			if (*g.board)[i][j] == 1 && (neighbors == 2 || neighbors == 3) {
				newBoard[i][j] = 1
			} else if (*g.board)[i][j] == 0 && neighbors == 3 {
				newBoard[i][j] = 1
			} else {
				newBoard[i][j] = 0
			}
		}
	}
	return &newBoard
}

func main() {
	width, height := getSize()
	game := newGame(width, height)
	game.beforeGame()
	game.draw()
	for {
		game.tick()
	}
}
